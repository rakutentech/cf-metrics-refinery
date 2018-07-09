package cli

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/pkg/errors"
	"github.com/rakutentech/cf-metrics-refinery/debug"
	"github.com/rakutentech/cf-metrics-refinery/enricher"
	"github.com/rakutentech/cf-metrics-refinery/input"
	"github.com/rakutentech/cf-metrics-refinery/output"
	"github.com/rakutentech/cf-metrics-refinery/transformer"
)

func TestConfigParse(t *testing.T) {
	CFMR_CF_API := ""
	CFMR_CF_USER := ""
	CFMR_CF_PASSWORD := ""
	CFMR_CF_TIMEOUT := "1m"
	CFMR_CF_SKIPSSLVALIDATION := "false"
	CFMR_CF_RESULTSPERPAGE := "50"
	CFMR_INFLUXDB_USERNAME := ""
	CFMR_INFLUXDB_PASSWORD := ""
	CFMR_INFLUXDB_ADDR := ""
	CFMR_INFLUXDB_TIMEOUT := "1m"
	CFMR_INFLUXDB_DATABASE := ""
	CFMR_INFLUXDB_RETENTIONPOLICY := ""
	CFMR_INFLUXDB_SKIPSSLVALIDATION := "false"
	CFMR_INFLUXDB_INFLUXPINGTIMEOUT := "5s"
	CFMR_BATCHER_FLUSHINTERVAL := "3s"
	CFMR_BATCHER_FLUSHMESSAGES := "5000"
	CFMR_KAFKA_ZOOKEEPERS := ""
	CFMR_KAFKA_TOPICS := ","
	CFMR_KAFKA_CONSUMERGROUP := ""
	CFMR_KAFKA_PROCESSINGTIMEOUT := "1m"
	CFMR_KAFKA_OFFSETNEWEST := "false"
	CFMR_SERVER_PORT := "8080"
	CFMR_METADATAREFRESH := "10m"
	CFMR_METADATAEXPIRE := "3m"
	CFMR_METADATAEXPIRECHECK := "1m"
	CFMR_NEGATIVECACHEEXPIRE := "20m"
	CFMR_NEGATIVECACHEEXPIRECHECK := "3m"

	// Convert string to int
	BATCHER_FLUSHMESSAGES, _ := strconv.Atoi(CFMR_BATCHER_FLUSHMESSAGES)
	CF_RESULTSPERPAGE, _ := strconv.Atoi(CFMR_CF_RESULTSPERPAGE)
	// Convert string to bool
	CF_SKIPSSLVALIDATION := false
	if CFMR_CF_SKIPSSLVALIDATION == "true" {
		CF_SKIPSSLVALIDATION = true
	}
	KAFKA_OFFSETNEWEST := false
	if CFMR_KAFKA_OFFSETNEWEST == "true" {
		KAFKA_OFFSETNEWEST = true
	}
	INFLUXDB_SKIPSSLVALIDATION := false
	if CFMR_INFLUXDB_SKIPSSLVALIDATION == "true" {
		INFLUXDB_SKIPSSLVALIDATION = true
	}
	// Convert string to time.Duration
	CF_TIMEOUT, _ := time.ParseDuration(CFMR_CF_TIMEOUT)
	INFLUXDB_TIMEOUT, _ := time.ParseDuration(CFMR_INFLUXDB_TIMEOUT)
	INFLUXDB_INFLUXPINGTIMEOUT, _ := time.ParseDuration(CFMR_INFLUXDB_INFLUXPINGTIMEOUT)
	BATCHER_FLUSHINTERVAL, _ := time.ParseDuration(CFMR_BATCHER_FLUSHINTERVAL)
	KAFKA_PROCESSINGTIMEOUT, _ := time.ParseDuration(CFMR_KAFKA_PROCESSINGTIMEOUT)
	METADATAREFRESH, _ := time.ParseDuration(CFMR_METADATAREFRESH)
	METADATAEXPIRE, _ := time.ParseDuration(CFMR_METADATAEXPIRE)
	METADATAEXPIRECHECK, _ := time.ParseDuration(CFMR_METADATAEXPIRECHECK)
	NEGATIVECACHEEXPIRE, _ := time.ParseDuration(CFMR_NEGATIVECACHEEXPIRE)
	NEGATIVECACHEEXPIRECHECK, _ := time.ParseDuration(CFMR_NEGATIVECACHEEXPIRECHECK)

	cfConfig := enricher.ConfigCF{
		API:               CFMR_CF_API,
		User:              CFMR_CF_USER,
		Password:          CFMR_CF_PASSWORD,
		Timeout:           CF_TIMEOUT,
		SkipSSLValidation: CF_SKIPSSLVALIDATION,
		ResultsPerPage:    CF_RESULTSPERPAGE,
	}
	influxDBConfig := output.ConfigInfluxDB{
		Username:          CFMR_INFLUXDB_USERNAME,
		Password:          CFMR_INFLUXDB_PASSWORD,
		SkipSSLValidation: INFLUXDB_SKIPSSLVALIDATION,
		Addr:              CFMR_INFLUXDB_ADDR,
		Timeout:           INFLUXDB_TIMEOUT,
		Database:          CFMR_INFLUXDB_DATABASE,
		RetentionPolicy:   CFMR_INFLUXDB_RETENTIONPOLICY,
		InfluxPingTimeout: INFLUXDB_INFLUXPINGTIMEOUT,
	}
	batcherConfig := output.ConfigBatcher{
		FlushInterval: BATCHER_FLUSHINTERVAL,
		FlushMessages: BATCHER_FLUSHMESSAGES,
	}
	kafkaConfig := input.ConfigKafka{
		Zookeepers:        CFMR_KAFKA_ZOOKEEPERS,
		Topics:            strings.Split(CFMR_KAFKA_TOPICS, ","),
		ConsumerGroup:     CFMR_KAFKA_CONSUMERGROUP,
		ProcessingTimeout: KAFKA_PROCESSINGTIMEOUT,
		OffsetNewest:      KAFKA_OFFSETNEWEST,
	}
	serverConfig := debug.ConfigServer{
		Port: CFMR_SERVER_PORT,
	}
	wantConfig := Config{
		CF:                       cfConfig,
		InfluxDB:                 influxDBConfig,
		Batcher:                  batcherConfig,
		Kafka:                    kafkaConfig,
		Server:                   serverConfig,
		MetadataRefresh:          METADATAREFRESH,
		MetadataExpire:           METADATAEXPIRE,
		MetadataExpireCheck:      METADATAEXPIRECHECK,
		NegativeCacheExpire:      NEGATIVECACHEEXPIRE,
		NegativeCacheExpireCheck: NEGATIVECACHEEXPIRECHECK,
	}

	os.Clearenv()
	os.Setenv("CFMR_CF_API", CFMR_CF_API)
	os.Setenv("CFMR_CF_USER", CFMR_CF_USER)
	os.Setenv("CFMR_CF_PASSWORD", CFMR_CF_PASSWORD)
	os.Setenv("CFMR_CF_TIMEOUT", CFMR_CF_TIMEOUT)
	os.Setenv("CFMR_CF_SKIPSSLVALIDATION", CFMR_CF_SKIPSSLVALIDATION)
	os.Setenv("CFMR_INFLUXDB_USERNAME", CFMR_INFLUXDB_USERNAME)
	os.Setenv("CFMR_INFLUXDB_PASSWORD", CFMR_INFLUXDB_PASSWORD)
	os.Setenv("CFMR_INFLUXDB_SKIPSSLVALIDATION", CFMR_INFLUXDB_SKIPSSLVALIDATION)
	os.Setenv("CFMR_INFLUXDB_ADDR", CFMR_INFLUXDB_ADDR)
	os.Setenv("CFMR_INFLUXDB_TIMEOUT", CFMR_INFLUXDB_TIMEOUT)
	os.Setenv("CFMR_INFLUXDB_DATABASE", CFMR_INFLUXDB_DATABASE)
	os.Setenv("CFMR_INFLUXDB_RETENTIONPOLICY", CFMR_INFLUXDB_RETENTIONPOLICY)
	os.Setenv("CFMR_INFLUXDB_INFLUXPINGTIMEOUT", CFMR_INFLUXDB_INFLUXPINGTIMEOUT)
	os.Setenv("CFMR_BATCHER_FLUSHINTERVAL", CFMR_BATCHER_FLUSHINTERVAL)
	os.Setenv("CFMR_BATCHER_FLUSHMESSAGES", CFMR_BATCHER_FLUSHMESSAGES)
	os.Setenv("CFMR_KAFKA_ZOOKEEPERS", CFMR_KAFKA_ZOOKEEPERS)
	os.Setenv("CFMR_KAFKA_TOPICS", CFMR_KAFKA_TOPICS)
	os.Setenv("CFMR_KAFKA_CONSUMERGROUP", CFMR_KAFKA_CONSUMERGROUP)
	os.Setenv("CFMR_KAFKA_PROCESSINGTIMEOUT", CFMR_KAFKA_PROCESSINGTIMEOUT)
	os.Setenv("CFMR_KAFKA_OFFSETNEWEST", CFMR_KAFKA_OFFSETNEWEST)
	os.Setenv("CFMR_SERVER_PORT", CFMR_SERVER_PORT)
	os.Setenv("CFMR_METADATAREFRESH", CFMR_METADATAREFRESH)
	os.Setenv("CFMR_METADATAEXPIRE", CFMR_METADATAEXPIRE)
	os.Setenv("CFMR_METADATAEXPIRECHECK", CFMR_METADATAEXPIRECHECK)
	os.Setenv("CFMR_NEGATIVECACHEEXPIRE", CFMR_NEGATIVECACHEEXPIRE)
	os.Setenv("CFMR_NEGATIVECACHEEXPIRECHECK", CFMR_NEGATIVECACHEEXPIRECHECK)

	// Test with the correct environment variable
	c, err := ConfigParse()
	if !reflect.DeepEqual(c, &wantConfig) || err != nil {
		t.Fatalf("TestConfigParse: expected %v, got %v", &wantConfig, c)
	}

	// Test the environment variable with the prefix are set that we don't want to parse
	os.Setenv("CFMR_FOR_TESTING_CONFIG", "for testing config function")
	ec, err := ConfigParse()
	if ec != nil || err == nil {
		t.Fatalf("TestConfigParse: expected nil, got %v", ec)
	}
}

func TestConfigUsage(t *testing.T) {
	var usage bytes.Buffer
	if err := ConfigUsage(&usage); err != nil {
		t.Fatalf("TestConfigUsage: %s", err)
	}
}

func TestName(t *testing.T) {
	cli := &CLI{}
	if cli.Name() != appName {
		t.Fatalf("TestName: expected %s, got %s", appName, cli.Name())
	}
}

func TestVersion(t *testing.T) {
	cli := &CLI{}
	if cli.Version() != version {
		t.Fatalf("TestVersion: expected %s, got %s", version, cli.Version())
	}
}

const LogMsg = `{
	"origin": "rep",
	"eventType": 5,
	"timestamp": 123456789012345678,
	"job": "cell",
	"index": "0",
	"ip": "192.168.0.50",
	"logMessage": {
		"message": "aGVsbG8gd29ybGQK",
		"message_type": 0,
		"timestamp": 123456789012345000,
		"app_id": "00000000-0000-0000-0000-000000000000",
		"source_type": "APP",
		"source_instance": "1"
	}
}`

var appMeta = enricher.AppMetadata{
	App:       "app",
	AppGUID:   "00000000-0000-0000-0000-000000000000",
	Space:     "space",
	SpaceGUID: "10000000-0000-0000-0000-000000000000",
	Org:       "org",
	OrgGUID:   "20000000-0000-0000-0000-000000000000",
}

func mockEvent(logMsg string) *events.Envelope {
	var e events.Envelope
	if err := json.Unmarshal([]byte(logMsg), &e); err != nil {
		return &events.Envelope{}
	}
	return &e
}

type mockReader struct {
	Envelope *transformer.Envelope
	Err      error
	l        sync.Mutex
}

func (mr *mockReader) Read() (*transformer.Envelope, error) {
	mr.l.Lock()
	defer mr.l.Unlock()

	if mr.Err != nil {
		return nil, mr.Err
	}

	envelope := &transformer.Envelope{Event: mockEvent(LogMsg)}
	mr.Envelope = envelope
	return envelope, nil
}

type mockEnricher struct {
	AppMeta enricher.AppMetadata
	Err     error
	l       sync.Mutex
}

func (me *mockEnricher) GetAppMetadata(appGUID string) (enricher.AppMetadata, error) {
	me.l.Lock()
	defer me.l.Unlock()

	if me.Err != nil {
		return enricher.AppMetadata{}, me.Err
	}

	me.AppMeta = appMeta
	return appMeta, nil
}

type mockWriteAsync struct {
	Envs []*transformer.Envelope
	Err  error
	l    sync.Mutex
}

func (mwa *mockWriteAsync) WriteAsync(envs ...*transformer.Envelope) error {
	mwa.l.Lock()
	defer mwa.l.Unlock()

	if mwa.Err == nil {
		mwa.Envs = envs
	} else {
		var emptyEnvs []*transformer.Envelope
		mwa.Envs = emptyEnvs
	}
	return mwa.Err
}

func (mwa *mockWriteAsync) Flush() error {
	return mwa.Err
}

func TestProcess_ReadFail(t *testing.T) {
	cli := &CLI{}
	mr := &mockReader{}
	me := &mockEnricher{}
	mwa := &mockWriteAsync{}
	s := &debug.Stats{}

	mr.Err = errors.New("mock error for reading message")

	err := cli.Process(mr, me, mwa, s)
	if err == nil {
		t.Fatal("TestProcess_ReadFail: expected error, got nil")
	}

	var wantEnvelope *transformer.Envelope
	if !reflect.DeepEqual(wantEnvelope, mr.Envelope) {
		t.Fatalf("TestProcess_ReadFail: expected Envelope %v, got %v", wantEnvelope, mr.Envelope)
	}

	wantConsume := uint64(0)
	if s.Consume != wantConsume {
		t.Fatalf("TestProcess_ReadFail: expected Consume %v, got %v", wantConsume, s.Consume)
	}
}

func TestProcess_WriteAsyncFail(t *testing.T) {
	cli := &CLI{}
	mr := &mockReader{}
	me := &mockEnricher{}
	mwa := &mockWriteAsync{}
	s := &debug.Stats{}

	mwa.Err = errors.New("mock error for WriteAsync")
	err := cli.Process(mr, me, mwa, s)
	if err == nil {
		t.Fatal("TestProcess_WriteAsyncFail: expected error, got nil")
	}

	wantEnvelope := &transformer.Envelope{Event: mockEvent(LogMsg), Meta: appMeta}
	if !reflect.DeepEqual(wantEnvelope, mr.Envelope) {
		t.Fatalf("TestProcess_WriteAsyncFail: expected Envelope %v, got %v", wantEnvelope, mr.Envelope)
	}

	wantAppMeta := appMeta
	if !reflect.DeepEqual(wantAppMeta, me.AppMeta) {
		t.Fatalf("TestProcess_WriteAsyncFail: expected AppMetaData %v, got %v", wantAppMeta, me.AppMeta)
	}

	var wantEnvs []*transformer.Envelope
	if !reflect.DeepEqual(wantEnvs, mwa.Envs) {
		t.Fatalf("TestProcess_WriteAsyncFail: expected []*transformer.Envelope %v, got %v", wantEnvs, mwa.Envs)
	}
}

func TestProcess_EnrichFail(t *testing.T) {
	cli := &CLI{Logger: log.New(os.Stdout, "[Testing]", log.LstdFlags)}
	mr := &mockReader{}
	me := &mockEnricher{}
	mwa := &mockWriteAsync{}
	s := &debug.Stats{}

	go func() {
		me.l.Lock()
		defer me.l.Unlock()

		mr.l.Lock()
		defer mr.l.Unlock()

		wantAppMeta := appMeta
		if !reflect.DeepEqual(wantAppMeta, me.AppMeta) {
			t.Fatalf("TestProcess_EnrichFail: expected AppMetaData %v, got %v", wantAppMeta, me.AppMeta)
		}

		me.Err = errors.New("mock error for enriching")
		mr.Err = errors.New("mock error for reading message")
	}()
	err := cli.Process(mr, me, mwa, s)
	if err == nil {
		t.Fatal("TestProcess_EnrichFail: expected error, got nil")
	}
}

func TestProcess_Success(t *testing.T) {
	cli := &CLI{}
	mr := &mockReader{}
	me := &mockEnricher{}
	mwa := &mockWriteAsync{}
	s := &debug.Stats{}

	go func() {
		mwa.l.Lock()
		defer mwa.l.Unlock()

		var wantEnvs []*transformer.Envelope
		wantEnvs = append(wantEnvs, &transformer.Envelope{Event: mockEvent(LogMsg), Meta: appMeta})
		if !reflect.DeepEqual(wantEnvs, mwa.Envs) {
			t.Fatalf("TestProcess_Success: expected []*transformer.Envelope %v, got %v", wantEnvs, mwa.Envs)
		}

		mwa.Err = errors.New("mock error for WriteAsync")
	}()

	err := cli.Process(mr, me, mwa, s)
	if err == nil {
		t.Fatal("TestProcess_Success: expected error, got nil")
	}

	wantEnvelope := &transformer.Envelope{Event: mockEvent(LogMsg), Meta: appMeta}
	if !reflect.DeepEqual(wantEnvelope, mr.Envelope) {
		t.Fatalf("TestProcess_Success: expected Envelope %v, got %v", wantEnvelope, mr.Envelope)
	}

	wantAppMeta := appMeta
	if !reflect.DeepEqual(wantAppMeta, me.AppMeta) {
		t.Fatalf("TestProcess_Success: expected AppMetaData %v, got %v", wantAppMeta, me.AppMeta)
	}

	var wantEmptyEnvs []*transformer.Envelope
	if !reflect.DeepEqual(wantEmptyEnvs, mwa.Envs) {
		t.Fatalf("TestProcess_WriteAsyncFail: expected empty []*transformer.Envelope %v, got %v", wantEmptyEnvs, mwa.Envs)
	}
}
