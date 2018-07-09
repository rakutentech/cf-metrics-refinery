package output

import "github.com/rakutentech/cf-metrics-refinery/transformer"

type Committer struct { // SyncWriter
	parent SyncWriter
	cb     CommitCallback
}

type CommitCallback func([]*transformer.Envelope) error

func NewCommitter(parent SyncWriter, cb CommitCallback) *Committer {
	return &Committer{parent, cb}
}

func (c *Committer) Write(envs ...*transformer.Envelope) error {
	err := c.parent.Write(envs...)
	if err == nil {
		err = c.cb(envs)
	}
	return err
}
