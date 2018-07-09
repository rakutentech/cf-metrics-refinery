package input

import "github.com/rakutentech/cf-metrics-refinery/transformer"

type Reader interface {
	Read() (*transformer.Envelope, error)
}
