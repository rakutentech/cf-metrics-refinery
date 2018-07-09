package output

import (
	"github.com/rakutentech/cf-metrics-refinery/transformer"
)

// AsyncWriter is the writer interface for outputs where the
// implementation is stateful
type AsyncWriter interface {
	// WriteAsync accepts one or messages for writing. If it returns
	// non-nil error the AsyncWriter (and the output itself) is in an
	// undefined state and must not be used anymore (i.e. no further
	// calls to WriteAsync or Flush should be made).
	// Implementations of WriteAsync should not assume anything about
	// how frequently Flush is called in comparison to WriteAsync.
	WriteAsync(...*transformer.Envelope) error

	// Flush ensures that all messages passed to WriteAsync before the
	// call to Flush have been flushed to the output. If it returns
	// non-nil error the AsyncWriter (and the output itself) is in an
	// undefined state and must not be used anymore (i.e. no further
	// calls to WriteAsync or Flush should be made).
	// Implementations of Flush should not assume anything about how
	// frequently WriteAsync is called in comparison to Flush.
	Flush() error
}

// SyncWriter is the writer instance for outputs where the writer
// implementation is stateless.
type SyncWriter interface {
	// Write writes one or messages to the output. If it returns non-nil
	// error it is undefined how many messages have been written to the
	// output.
	Write(...*transformer.Envelope) error
}
