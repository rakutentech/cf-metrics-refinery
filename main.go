package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/logutils"
	"github.com/rakutentech/cf-metrics-refinery/cli"
)

func main() {
	flags := flag.NewFlagSet((&cli.CLI{}).Name(), flag.ExitOnError) // FIXME: ugly
	logLevel := flags.String("log-level", "INFO", "")
	flags.SetOutput(os.Stderr)
	flags.Usage = func() {
		err := cli.ConfigUsage(os.Stderr)
		if err != nil {
			log.Fatal("[ERROR] Failed to provide Config usage: ", err)
		}
	}
	flags.Parse(os.Args[1:])

	cliConfig, err := cli.ConfigParse()
	if err != nil {
		log.Fatal("[ERROR] Failed to parse env configurations: ", err)
	}

	app := &cli.CLI{
		OutStream: os.Stdout,
		ErrStream: os.Stderr,
		Conf:      cliConfig,
	}

	// Setup logger with level Filtering
	logger := log.New(&logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "ERROR"},
		MinLevel: (logutils.LogLevel)(strings.ToUpper(*logLevel)),
		Writer:   app.OutStream,
	}, "", log.LstdFlags)
	logger.Printf("[INFO] %s version %s", app.Name(), app.Version())
	logger.Printf("[DEBUG] LogLevel: %s", *logLevel)

	os.Exit(app.Run(logger))
}
