package debug

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	defaultInstanceIndex int = 0
	EnvCFInstanceIndex       = "CF_INSTANCE_INDEX"
)

type StatsType int

const (
	Consume    StatsType = iota // messages received
	Enrich                      // messages enriched
	EnrichFail                  // messages failed to be enriched
	WriteAsync                  // points added to Influxdb batch
	Write                       // points written to Influxdb
	CFFail                      // CF API lookup failure
)

// Stats stores various stats infomation
type Stats struct {
	l                  sync.Mutex
	Consume            uint64    `json:"consume"`
	ConsumePerSec      uint64    `json:"consume_per_sec"`
	Enrich             uint64    `json:"enrich"`
	EnrichPerSec       uint64    `json:"enrich_per_sec"`
	EnrichFail         uint64    `json:"enrichfail"`
	EnrichFailPerSec   uint64    `json:"enrichfail_per_sec"`
	WriteAsync         uint64    `json:"writeasync"`
	WriteAsyncPerSec   uint64    `json:"writeasync_per_sec"`
	Write              uint64    `json:"write"`
	WritePerSec        uint64    `json:"write_per_sec"`
	CFFail             uint64    `json:"cffail"`
	CFFailPerSec       uint64    `json:"cffail_per_sec"`
	LastConsumeTime    time.Time `json:"last_consume_time"`
	LastEnrichTime     time.Time `json:"last_enrich_time"`
	LastEnrichFailTime time.Time `json:"last_enrich_fail_time"`
	LastWriteAsyncTime time.Time `json:"last_writeasync_time"`
	LastWriteTime      time.Time `json:"last_write_time"`
	LastCFFailTime     time.Time `json:"last_cffail_time"`
	// InstanceIndex is the index for cf-metrics-refinery instance.
	// This is used to identify stats from different instances.
	// By default, it's defaultInstanceIndex
	InstanceIndex int `json:"instance_index"`
}

func NewStats() *Stats {
	s := &Stats{}
	if idx, err := strconv.Atoi(os.Getenv(EnvCFInstanceIndex)); err == nil {
		s.InstanceIndex = idx
	}
	return s
}

func (s *Stats) PerSec() {
	var lastConsume, lastEnrich, lastEnrichFail, lastWriteAsync, lastWrite, lastCFFail uint64
	for range time.Tick(1 * time.Second) {

		s.l.Lock()

		s.ConsumePerSec = s.Consume - lastConsume
		s.EnrichPerSec = s.Enrich - lastEnrich
		s.EnrichFailPerSec = s.EnrichFail - lastEnrichFail
		s.WriteAsyncPerSec = s.WriteAsync - lastWriteAsync
		s.WritePerSec = s.Write - lastWrite
		s.CFFailPerSec = s.CFFail - lastCFFail

		lastConsume = s.Consume
		lastEnrich = s.Enrich
		lastEnrichFail = s.EnrichFail
		lastWriteAsync = s.WriteAsync
		lastWrite = s.Write
		lastCFFail = s.CFFail

		s.l.Unlock()
	}
}

func (s *Stats) Inc(statsType StatsType, value int) {
	s.l.Lock()

	v := uint64(value)
	now := time.Now()
	switch statsType {
	case Consume:
		s.Consume += v
		s.LastConsumeTime = now
	case Enrich:
		s.Enrich += v
		s.LastEnrichTime = now
	case EnrichFail:
		s.EnrichFail += v
		s.LastEnrichFailTime = now
	case WriteAsync:
		s.WriteAsync += v
		s.LastWriteAsyncTime = now
	case Write:
		s.Write += v
		s.LastWriteTime = now
	case CFFail:
		s.CFFail += v
		s.LastCFFailTime = now
	default:
		s.l.Unlock()
		panic(fmt.Sprintf("statsType is %s, not expected.", statsType))
	}
	s.l.Unlock()
}

func (s *Stats) Json() ([]byte, error) {
	s.l.Lock()
	defer s.l.Unlock()

	return json.Marshal(s)
}
