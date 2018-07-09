package output

import (
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/rakutentech/cf-metrics-refinery/transformer"
)

type recordWriter struct {
	rec [][]*transformer.Envelope
	err error
}

func (w *recordWriter) Write(e ...*transformer.Envelope) error {
	w.rec = append(w.rec, e)
	return w.err
}

func TestBatcher(t *testing.T) {
	w := &recordWriter{}
	b := NewBatcher(w, ConfigBatcher{100 * time.Millisecond, 2})

	e := b.WriteAsync(&transformer.Envelope{})
	if e != nil {
		t.Fatalf("unexpected error %v", e)
	}
	if len(w.rec) != 0 {
		t.Fatal("unexpected write")
	}

	e = b.WriteAsync(&transformer.Envelope{})
	if e != nil {
		t.Fatalf("unexpected error %v", e)
	}
	if len(w.rec) != 1 && len(w.rec[0]) != 2 {
		t.Fatal("missing write")
	}

	e = b.WriteAsync(&transformer.Envelope{})
	if e != nil {
		t.Fatalf("unexpected error %v", e)
	}
	if len(w.rec) != 1 && len(w.rec[0]) != 2 {
		t.Fatal("unexpected write")
	}

	time.Sleep(100 * time.Millisecond)
	if len(w.rec) != 2 && len(w.rec[0]) != 2 && len(w.rec[1]) != 1 {
		t.Fatal("missing write")
	}
}

func TestBatcherFlush(t *testing.T) {
	w := &recordWriter{}
	b := NewBatcher(w, ConfigBatcher{1000 * time.Millisecond, 1000})

	e := b.WriteAsync(&transformer.Envelope{})
	if e != nil {
		t.Fatalf("unexpected error %v", e)
	}
	if len(w.rec) != 0 {
		t.Fatal("unexpected write")
	}

	e = b.Flush()
	if e != nil {
		t.Fatalf("unexpected error %v", e)
	}
	if len(w.rec) != 1 && len(w.rec[0]) != 1 {
		t.Fatal("missing write")
	}
}

func TestBatcherWriteFailure(t *testing.T) {
	exp := errors.New("oh god why")
	w := &recordWriter{err: exp}
	b := NewBatcher(w, ConfigBatcher{500 * time.Millisecond, 2})

	e := b.WriteAsync(&transformer.Envelope{})
	if e != nil {
		t.Fatalf("unexpected error %v", e)
	}
	if len(w.rec) != 0 {
		t.Fatal("unexpected write")
	}

	e = b.WriteAsync(&transformer.Envelope{})
	if errors.Cause(e) != exp {
		t.Fatalf("unexpected error %v", e)
	}
	if len(w.rec) != 1 && len(w.rec[0]) != 2 {
		t.Fatal("missing write")
	}
}

func TestBatcherFlushFailure(t *testing.T) {
	exp := errors.New("oh god why")
	w := &recordWriter{err: exp}
	b := NewBatcher(w, ConfigBatcher{500 * time.Millisecond, 2})

	e := b.WriteAsync(&transformer.Envelope{})
	if e != nil {
		t.Fatalf("unexpected error %v", e)
	}
	if len(w.rec) != 0 {
		t.Fatal("unexpected write")
	}

	e = b.Flush()
	if errors.Cause(e) != exp {
		t.Fatalf("unexpected error %v", e)
	}
	if len(w.rec) != 1 && len(w.rec[0]) != 1 {
		t.Fatal("missing write")
	}
}
