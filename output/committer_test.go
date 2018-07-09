package output

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/rakutentech/cf-metrics-refinery/transformer"
)

type dumbWriter struct { // SyncWriter
	err error
}

func (w *dumbWriter) Write(_ ...*transformer.Envelope) error {
	return w.err
}

// if parent returns ok, callback should be called with all envelopes;
// if callback returns ok, we should get no error
func TestCommitterSuccess(t *testing.T) {
	e := []*transformer.Envelope{
		&transformer.Envelope{},
		&transformer.Envelope{},
		&transformer.Envelope{},
	}

	ok := false
	c := NewCommitter(&dumbWriter{}, func(envs []*transformer.Envelope) error {
		ok = len(envs) == len(e)
		for i := range e {
			ok = ok && (e[i] == envs[i])
		}
		return nil
	})

	err := c.Write(e...)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if !ok {
		t.Fatal("commit callback not invoked correctly")
	}
}

// if writer returns ok but commit fails, we should get the commit error
func TestCommitterFailure(t *testing.T) {
	e := []*transformer.Envelope{
		&transformer.Envelope{},
		&transformer.Envelope{},
		&transformer.Envelope{},
	}

	exp := errors.New("OH MY GOD EVERYTHING IS BURNING")
	c := NewCommitter(&dumbWriter{}, func(envs []*transformer.Envelope) error {
		return exp
	})

	err := c.Write(e...)
	if err != exp {
		t.Fatalf("unexpected error %v", err)
	}
}

// if writer returns error the callback should not be called and we should
// get the writer error
func TestCommitterParentFailure(t *testing.T) {
	e := []*transformer.Envelope{
		&transformer.Envelope{},
		&transformer.Envelope{},
		&transformer.Envelope{},
	}

	exp := errors.New("OH MY GOD EVERYTHING IS BURNING")
	ok := true
	c := NewCommitter(&dumbWriter{exp}, func(envs []*transformer.Envelope) error {
		ok = false
		return nil
	})

	err := c.Write(e...)
	if err != exp {
		t.Fatalf("unexpected error %v", err)
	}
	if !ok {
		t.Fatal("commit callback invoked unexpectedly")
	}
}
