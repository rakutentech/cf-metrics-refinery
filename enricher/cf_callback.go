package enricher

type CfCallback struct {
	parent Enricher
	cb     Callback
}

type Callback func(err error)

func NewCfCallback(parent Enricher, cb Callback) *CfCallback {
	return &CfCallback{parent, cb}
}

func (c *CfCallback) GetAppMetadata(app_guid string) (AppMetadata, error) {
	md, err := c.parent.GetAppMetadata(app_guid)
	c.cb(err)
	return md, err
}
