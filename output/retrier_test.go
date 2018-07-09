package output

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/rakutentech/cf-metrics-refinery/transformer"
)

type failingWriter struct {
	failures int
}

func (e *failingWriter) Write(_ ...*transformer.Envelope) error {
	if e.failures > 0 {
		e.failures--
		return errFailingWriter
	}
	return nil
}

var errFailingWriter = errors.New("so much fail")

func TestRetrier_Write(t *testing.T) {
	type fields struct {
		parent SyncWriter
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"fail always", fields{&failingWriter{1 << 30}}, true},
		{"fail 3 times", fields{&failingWriter{3}}, true},
		{"fail 2 times", fields{&failingWriter{2}}, false},
		{"fail 0 times", fields{&failingWriter{0}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewRetrier(tt.fields.parent)
			err := e.Write()
			if (tt.wantErr && err != errFailingWriter) || (!tt.wantErr && err != nil) {
				t.Errorf("Retrier.GetAppMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
