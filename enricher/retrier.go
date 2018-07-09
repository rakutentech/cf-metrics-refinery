package enricher

type Retrier struct {
	parent  Enricher
	retries int
}

// TODO: change retries to be configurable
func NewRetrier(e Enricher) Enricher {
	return &Retrier{parent: e, retries: 2}
}

func (e *Retrier) GetAppMetadata(app_guid string) (AppMetadata, error) {
	md, err := e.parent.GetAppMetadata(app_guid)
	// TODO: be smarter about retries: e.g. if CAPI says the app does not exist
	// there is no point in retrying
	for i := 0; i < e.retries && err != nil; i++ {
		md, err = e.parent.GetAppMetadata(app_guid)
	}
	return md, err
}
