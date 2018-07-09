package enricher

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/pkg/errors"
)

type CFClient struct {
	c   *cfclient.Client
	cfg ConfigCF
}

type ConfigCF struct {
	API               string        `required:"true" desc:"URL of the Cloud Foundry API endpoint"`                          // CFMR_CF_API
	User              string        `required:"true" desc:"Username for the Cloud Foundry API"`                             // CFMR_CF_USER
	Password          string        `required:"true" desc:"Password for the Cloud Foundry API"`                             // CFMR_CF_PASSWORD
	Timeout           time.Duration `default:"1m" desc:"Timeout for Cloud Foundry API requests"`                            // CFMR_CF_TIMEOUT
	SkipSSLValidation bool          `default:"false" desc:"Skip SSL certificate validation for Cloud Foundry API requests"` // CFMR_CF_SKIPSSLVALIDATION
	ResultsPerPage    int           `default:"50" desc:"Number of results per page to fetch from CF API"`                   // CFMR_CF_RESULTSPERPAGE
	Token             string        `desc:"Token for Cloud Foundry API"`                                                    // CFMR_CF_TOKEN
	ClientID          string        `desc:"Client ID for Cloud Foundry API"`                                                // CFMR_CF_CLIENTID
	ClientSecret      string        `desc:"Client secret for Cloud Foundry API"`                                            // CFMR_CF_CLIENTSECRET
	UserAgent         string        `ignored:"true"`
}

func NewCFClient(cfg ConfigCF) (*CFClient, error) {
	c, err := cfclient.NewClient(&cfclient.Config{
		ApiAddress: cfg.API,
		Username:   cfg.User,
		Password:   cfg.Password,
		UserAgent:  cfg.UserAgent,
		HttpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		SkipSslValidation: cfg.SkipSSLValidation,
		Token:             cfg.Token,
		ClientID:          cfg.ClientID,
		ClientSecret:      cfg.ClientSecret,
	})
	if err != nil {
		return nil, errors.Wrap(err, "creating cfclient")
	}

	if cfg.ResultsPerPage <= 0 {
		return nil, errors.Errorf("invalid value for ResultPerPage: %d", cfg.ResultsPerPage)
	}

	return &CFClient{c: c, cfg: cfg}, nil
}

// GetAppMetadata returns the metadata for the application with the specified GUID
func (e *CFClient) GetAppMetadata(appGUID string) (AppMetadata, error) {
	App, err := e.c.AppByGuid(appGUID)
	if err != nil {
		return AppMetadata{}, errors.Wrap(err, "getting app metadata")
	}

	Space, err := App.Space()
	if err != nil {
		return AppMetadata{}, errors.Wrap(err, "getting space metadata")
	}

	Org, err := Space.Org()
	if err != nil {
		return AppMetadata{}, errors.Wrap(err, "getting org metadata")
	}

	return AppMetadata{
		App:       App.Name,
		Space:     Space.Name,
		Org:       Org.Name,
		AppGUID:   App.Guid,
		SpaceGUID: Space.Guid,
		OrgGUID:   Org.Guid,
	}, nil
}

// GetRunningAppMetadata returns the metadata for all STARTED applications.
func (e *CFClient) GetRunningAppMetadata() ([]AppMetadata, error) {
	q := url.Values{}
	q.Set("results-per-page", strconv.Itoa(e.cfg.ResultsPerPage))

	orgs, err := e.c.ListOrgsByQuery(q)
	if err != nil {
		return nil, errors.Wrap(err, "listing all orgs")
	}

	spaces, err := e.c.ListSpacesByQuery(q)
	if err != nil {
		return nil, errors.Wrap(err, "listing all spaces")
	}

	// note: we don't need inline-relations-depth=2 because we manually
	// grab orgs and spaces above and join them below
	apps, err := e.c.ListAppsByQuery(q)
	if err != nil {
		return nil, errors.Wrap(err, "listing all apps")
	}

	return joinAppSpaceOrg(apps, spaces, orgs), nil
}

func joinAppSpaceOrg(apps []cfclient.App, spaces []cfclient.Space, orgs []cfclient.Org) []AppMetadata {
	orgmap := make(map[string]cfclient.Org, len(orgs))
	for _, org := range orgs {
		orgmap[org.Guid] = org
	}

	spacemap := make(map[string]cfclient.Space, len(spaces))
	for _, space := range spaces {
		spacemap[space.Guid] = space
	}

	allAppMetadata := make([]AppMetadata, 0, len(apps))
	for _, app := range apps {
		if app.State != "STARTED" {
			// TODO: https://github.com/cloudfoundry/cloud_controller_ng/issues/1109
			continue
		}
		if space, found := spacemap[app.SpaceGuid]; found {
			if org, found := orgmap[space.OrganizationGuid]; found {
				allAppMetadata = append(allAppMetadata, AppMetadata{
					App:       app.Name,
					Space:     space.Name,
					Org:       org.Name,
					AppGUID:   app.Guid,
					SpaceGUID: space.Guid,
					OrgGUID:   org.Guid,
				})
			}
		}
	}

	return allAppMetadata
}
