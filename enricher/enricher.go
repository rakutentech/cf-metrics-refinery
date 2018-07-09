package enricher

type Enricher interface {
	GetAppMetadata(app_guid string) (AppMetadata, error)
}

type AppMetadata struct {
	App       string
	Space     string
	Org       string
	AppGUID   string
	SpaceGUID string
	OrgGUID   string
}
