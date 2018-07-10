package cli

import (
	"io"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/rakutentech/cf-metrics-refinery/debug"
	"github.com/rakutentech/cf-metrics-refinery/enricher"
	"github.com/rakutentech/cf-metrics-refinery/input"
	"github.com/rakutentech/cf-metrics-refinery/output"
	"github.com/rakutentech/cf-metrics-refinery/transformer"
)

// Exit codes are int values that represent an exit code for a particular error.
const (
	ExitCodeOK    int = 0
	ExitCodeError int = 1 + iota

	envPrefix = "cfmr"
	appName   = "cf-metrics-refinery"
)

var version = "1.0.0-dev"

// CLI is the command line object
type CLI struct {
	// OutStream and ErrStream are the stdout and stderr to write message from the CLI.
	OutStream, ErrStream io.Writer
	Conf                 *Config
	//LogLevel string
	Logger *log.Logger
}

// Config is the root configuration structure
type Config struct {
	CF       enricher.ConfigCF
	InfluxDB output.ConfigInfluxDB
	Batcher  output.ConfigBatcher
	Kafka    input.ConfigKafka
	Server   debug.ConfigServer

	MetadataRefresh          time.Duration `default:"10m" desc:"How often to fetch a fresh copy of all metadata"`
	MetadataExpire           time.Duration `default:"3m" desc:"How long before metadata is considered expired"`
	MetadataExpireCheck      time.Duration `default:"1m" desc:"How often to check for expired metadata"`
	NegativeCacheExpire      time.Duration `default:"20m" desc:"How long before negative cache is considered expired"`
	NegativeCacheExpireCheck time.Duration `default:"3m" desc:"How often to check for expired negative cache"`
}

// ConfigParse parses the CFMR_* environment vars to extract the configuration
func ConfigParse() (*Config, error) {
	config := &Config{}
	if err := envconfig.CheckDisallowed(envPrefix, config); err != nil {
		return nil, err
	}
	if err := envconfig.Process(envPrefix, config); err != nil {
		return nil, err
	}
	return config, nil
}

// ConfigUsage lists the environment variables to be used for configuration
func ConfigUsage(w io.Writer) error {
	err := envconfig.Usagef(envPrefix, &Config{}, w, envconfig.DefaultTableFormat)
	return err
}

// Run invokes the CLI
func (cli *CLI) Run(logger *log.Logger) int {
	cli.Logger = logger

	if cli.Conf == nil {
		cli.Logger.Println("[ERROR] Configuration is missing")
		return ExitCodeError
	}

	userAgent := appName + "/" + version

	// Initialize stats collector
	stats := debug.NewStats()
	go stats.PerSec()

	// Set up http server to expose stats endpoints for debugging
	server, err := debug.NewServer(&cli.Conf.Server, stats, logger)
	if err != nil {
		cli.Logger.Println("[ERROR] New Http Server", err)
		return ExitCodeOK
	}
	srv := server.Start()
	defer srv.Shutdown(nil)

	// Build the enricher chain
	cli.Conf.CF.UserAgent = userAgent
	cfclient, cache, negativeCache, err := cli.EnricherChain(stats)
	if err != nil {
		cli.Logger.Println("[ERROR] Failed to build the enricher chain", err)
		return ExitCodeError
	}

	// Initial warmup
	cli.Logger.Print("[INFO] Warming up metadata cache")
	start := time.Now()
	appMetadataRunning, err := cfclient.GetRunningAppMetadata()
	if err != nil {
		cli.Logger.Println("[WARN] Failed to warmup metadata cache", err)
	}
	cache.(*enricher.MemLRUCache).Warmup(appMetadataRunning)
	cli.Logger.Printf("[INFO] Warming up metadata cache: %v", time.Since(start))

	// Build the input chain
	consumer, err := cli.InputChain()
	if err != nil {
		cli.Logger.Println("[ERROR] Failed to build input chain", err)
		return ExitCodeError
	}

	// Check consumer group errors
	go func() {
		cli.CGErrorsCheck(consumer)
	}()

	// Metadata cache eviction loop
	go func() {
		cli.CacheEvict(cache)
	}()

	// Metadata cache refresh loop
	go func() {
		cli.CacheRefresh(cache, negativeCache, cfclient)
	}()

	// Negative cache eviction loop
	go func() {
		cli.NegativeCacheEvict(negativeCache)
	}()

	// Build the output chain
	cli.Logger.Println("[INFO] Configured InfluxDB, db:", cli.Conf.InfluxDB.Database)
	cli.Conf.InfluxDB.UserAgent = userAgent
	batcher, err := cli.OutputChain(consumer, stats)
	if err != nil {
		cli.Logger.Println("[ERROR] Failed to build the output chain", err)
		return ExitCodeError
	}

	// Display stats periodically
	go func() {
		cli.StatsEmit(stats)
	}()

	done := make(chan struct{}, 1)

	// Main envelope processing loop
	go func() {
		cli.Logger.Println("[INFO] Started processing.")
		if err := cli.Process(consumer, cache, batcher, stats); err != nil {
			cli.Logger.Println(err)
		}

		done <- struct{}{}
	}()

	// Gracefully stop when receving a signal
	go func() {
		cli.TrapSignals(consumer)
	}()

	<-done
	cli.Logger.Println("[INFO] Finished processing.")

	return ExitCodeOK
}

func (_ *CLI) Name() string {
	return appName
}

func (_ *CLI) Version() string {
	return version
}

func (cli *CLI) Process(consumer input.Reader, cache enricher.Enricher, batcher output.AsyncWriter, stats *debug.Stats) error {
	for {
		// Read a message
		te, err := consumer.Read()
		if err != nil {
			return errors.Wrap(err, "[WARNING] Failed to consume from Kafka")
		}
		stats.Inc(debug.Consume, 1)

		// Enrich
		err = te.Enrich(cache)
		if err != nil {
			errAppNotFound := "CF-AppNotFound"
			errNoneGUID := "envelope does not contain an app GUID"
			if !strings.Contains(err.Error(), errAppNotFound) && !strings.Contains(err.Error(), errNoneGUID) {
				cli.Logger.Println("[WARN] Failed to enrich", te.Meta, err)
			}
			stats.Inc(debug.EnrichFail, 1)
			continue
		}
		stats.Inc(debug.Enrich, 1)

		// Write a message
		err = batcher.WriteAsync(te)
		if err != nil {
			// retry logic is implemented in the output chain: if we receive an error
			// the only thing we can do is abort (see AsyncWriter docs)
			return errors.Wrap(err, "[ERROR] Failed to write point to InfluxDB")
		}
		stats.Inc(debug.WriteAsync, 1)
	}
}

func (cli *CLI) InputChain() (*input.KafkaConsumer, error) {
	consumer, err := input.NewKafkaConsumer(&cli.Conf.Kafka)
	if err != nil {
		cli.Logger.Println("[ERROR] Failed to create consumer group", err)
	}
	return consumer, err
}

func (cli *CLI) EnricherChain(stats *debug.Stats) (*enricher.CFClient, enricher.Enricher, enricher.Enricher, error) {
	cfclient, err := enricher.NewCFClient(cli.Conf.CF)
	if err != nil {
		cli.Logger.Println("[ERROR] Failed to create CF API client", err)
		return &enricher.CFClient{}, nil, nil, err
	}
	retrier := enricher.NewRetrier(cfclient)
	cfCallback := enricher.NewCfCallback(retrier, func(err error) {
		if err != nil {
			stats.Inc(debug.CFFail, 1)
		}
	})
	negativeCache := enricher.NewNegativeMemLRUCache(cfCallback)
	cache := enricher.NewMemLRUCache(negativeCache)
	return cfclient, cache, negativeCache, nil
}

func (cli *CLI) OutputChain(consumer *input.KafkaConsumer, stats *debug.Stats) (output.AsyncWriter, error) {
	influx, err := output.NewInfluxDB(cli.Conf.InfluxDB)
	if err != nil {
		cli.Logger.Println("[ERROR] Failed to create InfluxDB output", err)
		return nil, err
	}

	err = influx.Ping(cli.Conf.InfluxDB.InfluxPingTimeout)
	if err != nil {
		cli.Logger.Println("[ERROR] Influxdb server is not up", err)
		time.Sleep(30)
		return nil, err
	}

	outputRetrier := output.NewRetrier(influx)
	committer := output.NewCommitter(outputRetrier, func(e []*transformer.Envelope) error {
		for _, m := range e {
			// TODO: commit-after-write is not covered by unit tests and it should be
			if err := consumer.CG.CommitUpto(m.Input.(*sarama.ConsumerMessage)); err != nil {
				return err
			}
		}
		stats.Inc(debug.Write, len(e))
		return nil
	})
	batcher := output.NewBatcher(committer, cli.Conf.Batcher)

	return batcher, nil
}

func (cli *CLI) CGErrorsCheck(consumer *input.KafkaConsumer) {
	for err := range consumer.CG.Errors() {
		// FIXME: this should be properly handled
		cli.Logger.Println("[ERROR] Kafka consumer group Error", err)
	}
}

func (cli *CLI) CacheEvict(cache enricher.Enricher) {
	for _ = range time.Tick(cli.Conf.MetadataExpireCheck) {
		cli.Logger.Print("[INFO] Expiring metadata cache")
		start := time.Now()
		cache.(*enricher.MemLRUCache).Expire(cli.Conf.MetadataExpire)
		cli.Logger.Printf("[INFO] Expiring metadata cache: %v", time.Since(start))
	}
}

func (cli *CLI) NegativeCacheEvict(negativeCache enricher.Enricher) {
	for _ = range time.Tick(cli.Conf.NegativeCacheExpireCheck) {
		cli.Logger.Print("[INFO] Expiring negative cache")
		start := time.Now()
		negativeCache.(*enricher.NegativeMemLRUCache).Expire(cli.Conf.NegativeCacheExpire)
		cli.Logger.Printf("[INFO] Expiring negative cache: %v", time.Since(start))
	}
}

func (cli *CLI) CacheRefresh(cache enricher.Enricher, negativeCache enricher.Enricher, cfclient *enricher.CFClient) {
	intvl := cli.Conf.MetadataRefresh.Seconds() * (rand.Float64() - 0.5) / 5 // ±10%
	for _ = range time.Tick(cli.Conf.MetadataRefresh + time.Duration(intvl*float64(time.Second))) {
		cli.Logger.Println("[INFO] Refreshing metadata cache")
		start := time.Now()
		appMetadataRunning, err := cfclient.GetRunningAppMetadata()
		if err != nil {
			cli.Logger.Println("[WARN] Failed to refresh metadata and negative cache", err)
		} else {
			cache.(*enricher.MemLRUCache).Warmup(appMetadataRunning)
			negativeCache.(*enricher.NegativeMemLRUCache).Warmup(appMetadataRunning)
			cli.Logger.Printf("[INFO] Refreshing metadata and negative cache: %v", time.Since(start))
		}
	}
}

func (cli *CLI) StatsEmit(stats *debug.Stats) {
	for _ = range time.Tick(1 * time.Minute) {
		status, err := stats.Json()
		if err != nil {
			cli.Logger.Printf("[ERROR] Stats Emit: %v\n", err)
		}
		cli.Logger.Printf("[INFO] Stats: %v\n", string(status))
	}
}

func (cli *CLI) TrapSignals(consumer *input.KafkaConsumer) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-c
	cli.Logger.Println("[INFO] Signal caught")

	cli.Logger.Println("[INFO] Closing input")
	if err := consumer.Close(); err != nil {
		cli.Logger.Println("[ERROR] Failed to close input", err)
	}

	// closing the input will eventually cause consumer.Read() to fail
	// (in the main processing loop)
}
