package output

import (
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rakutentech/cf-metrics-refinery/transformer"
)

type Batcher struct { // AsyncWriter
	parent SyncWriter

	cfg ConfigBatcher

	l          sync.Mutex
	firstWrite time.Time
	envelopes  []*transformer.Envelope
}

type ConfigBatcher struct {
	FlushInterval time.Duration `default:"3s" desc:"How often to flush pending events"`     // CFMR_BATCHER_FLUSHINTERVAL
	FlushMessages int           `default:"5000" desc:"How many messages to flush together"` //CFMR_BATCHER_FLUSHMESSAGES
}

func NewBatcher(parent SyncWriter, cfg ConfigBatcher) *Batcher {
	return &Batcher{parent: parent, cfg: cfg}
}

// Write point to batch and handle with both the time-based and size-based flush
func (b *Batcher) WriteAsync(envs ...*transformer.Envelope) error {
	b.l.Lock()
	defer b.l.Unlock()

	// FIXME: make time-based flushing independent of WriteAsync calls
	if len(b.envelopes) == 0 {
		b.firstWrite = time.Now()
	}
	shouldFlushTimeBased := time.Since(b.firstWrite) >= b.cfg.FlushInterval

	b.envelopes = append(b.envelopes, envs...)
	shouldFlushSizeBased := len(b.envelopes) >= b.cfg.FlushMessages

	if shouldFlushSizeBased || shouldFlushTimeBased {
		err := b.flushWithLock()
		if err != nil {
			return errors.Wrap(err, "flushing")
		}
	}

	return nil
}

func (b *Batcher) Flush() error {
	b.l.Lock()
	defer b.l.Unlock()
	return b.flushWithLock()
}

func (b *Batcher) flushWithLock() error {
	err := b.parent.Write(b.envelopes...)
	b.envelopes = nil
	return err
}
