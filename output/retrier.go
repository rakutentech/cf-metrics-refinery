package output

import "github.com/rakutentech/cf-metrics-refinery/transformer"

type Retrier struct { // SyncWriter
	parent  SyncWriter
	retries int
}

func NewRetrier(e SyncWriter) *Retrier {
	// TODO: change retries to be configurable
	return &Retrier{parent: e, retries: 2}
}

func (e *Retrier) Write(envs ...*transformer.Envelope) error {
	err := e.parent.Write(envs...)
	for i := 0; i < e.retries && err != nil; i++ {
		err = e.parent.Write(envs...)
	}
	return err
}
