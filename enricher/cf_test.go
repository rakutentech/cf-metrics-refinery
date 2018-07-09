package enricher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
)

var (
	mux      *http.ServeMux
	server   *httptest.Server
	cfClient *CFClient
)

const (
	// Case: AppGuid, SpaceGuid and OrgGuid exist
	appGuidOK   = "1a8a5b6e-6a0c-4d87-8c3a-eff7b6da8c7e"
	appNameOK   = "testApp"
	spaceGuidOK = "193a2c3e-cc9e-4518-9a0d-e74d18ce04c7"
	spaceNameOK = "testSpace"
	orgGuidOK   = "3f5e86a5-2af4-45d9-852d-d7433078e0d4"
	orgNameOK   = "testOrg"

	// Error Case 1: App Guid doesn't exist
	appGuidErr1 = "00000000-0000-0000-0000-000000000001"

	// Error Case 2: App Guid exists, Space Guid doesn't exist
	appGuidErr2   = "00000000-0000-0000-0000-000000000002"
	appNameErr2   = "testAppErr2"
	spaceGuidErr2 = "10000000-0000-0000-0000-000000000002"

	// Error Case 3: App Guid and Space Guid exist, Org Guid does not exist
	appGuidErr3   = "00000000-0000-0000-0000-000000000003"
	appNameErr3   = "testAppErr3"
	spaceGuidErr3 = "10000000-0000-0000-0000-000000000003"
	spaceNameErr3 = "testSpaceErr3"
	orgGuidErr3   = "20000000-0000-0000-0000-000000000003"

	// Case: App not started
	appGuidNotStarted = "00000000-0000-0000-0000-000000000004"
	appNameNotStarted = "testAppNotStarted"
)

func TestCFWarmupJoinerLogic(t *testing.T) {
	org := cfclient.Org{Guid: "org_guid", Name: "org_name"}
	space := cfclient.Space{Guid: "space_guid", Name: "space_name", OrganizationGuid: org.Guid}
	app1 := cfclient.App{Guid: "app1_guid", Name: "app1_name", SpaceGuid: space.Guid, State: "STARTED"}
	app2 := cfclient.App{Guid: "app2_guid", Name: "app2_name", SpaceGuid: space.Guid, State: "STARTED"}
	// app3 is STOPPED, so it won't show up in the results
	app3 := cfclient.App{Guid: "app3_guid", Name: "app3_name", SpaceGuid: space.Guid, State: "STOPPED"}
	// app4 is in a unknown space, so it won't show up in the results
	app4 := cfclient.App{Guid: "app4_guid", Name: "app4_name", SpaceGuid: "unknown", State: "STARTED"}
	// orgB and spaceB are not used by any app
	orgB := cfclient.Org{Guid: "orgB_guid", Name: "orgB_name"}
	spaceB := cfclient.Space{Guid: "spaceB_guid", Name: "spaceB_name", OrganizationGuid: orgB.Guid}

	apps := []cfclient.App{app1, app2, app3, app4}
	spaces := []cfclient.Space{space, spaceB}
	orgs := []cfclient.Org{org, orgB}

	res := joinAppSpaceOrg(apps, spaces, orgs)

	if len(res) != 2 {
		t.Fatal("unexpected number of app metadata")
	}
	if res[0].App != app1.Name || res[0].AppGUID != app1.Guid ||
		res[0].Space != space.Name || res[0].SpaceGUID != space.Guid ||
		res[0].Org != org.Name || res[0].OrgGUID != org.Guid {
		t.Fatalf("unexpected metadata for app1: %+v", res[0])
	}
	if res[1].App != app2.Name || res[1].AppGUID != app2.Guid ||
		res[1].Space != space.Name || res[1].SpaceGUID != space.Guid ||
		res[1].Org != org.Name || res[1].OrgGUID != org.Guid {
		t.Fatalf("unexpected metadata for app2: %+v", res[1])
	}
}

func setup() func() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	mux.HandleFunc("/v2/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		endP := cfclient.Endpoint{TokenEndpoint: server.URL}
		endPoint, _ := json.Marshal(endP)
		fmt.Fprint(w, string(endPoint))
	})

	// TokenEndpoint
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		oauthToken := make(map[string]interface{})
		oauthToken["access_token"] = "test"
		oauthToken["token_type"] = "test"
		oauthToken["refresh_token"] = "test"
		oauthToken["expires_in"] = 599
		oauthToken["scope"] = "test"
		oauthToken["jti"] = "test"

		token, _ := json.Marshal(oauthToken)
		fmt.Fprint(w, string(token))
	})

	cfClient, _ = NewCFClient(ConfigCF{
		API:               server.URL,
		User:              "test",
		Password:          "test",
		SkipSSLValidation: true,
		ResultsPerPage:    100,
	})

	return func() {
		server.Close()
	}
}

func TestCFGetAppMetadata(t *testing.T) {
	teardown := setup()
	defer teardown()

	// Mock API for Case: AppGuid, SpaceGuid and OrgGuid exist
	mux.HandleFunc("/v2/apps/"+appGuidOK, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		appR := cfclient.AppResource{Meta: cfclient.Meta{Guid: appGuidOK}, Entity: cfclient.App{Guid: appGuidOK, Name: appNameOK, SpaceURL: "/v2/spaces/" + spaceGuidOK}}
		appResource, _ := json.Marshal(appR)
		fmt.Fprint(w, string(appResource))
	})

	mux.HandleFunc("/v2/spaces/"+spaceGuidOK, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		spaceR := cfclient.SpaceResource{Meta: cfclient.Meta{Guid: spaceGuidOK}, Entity: cfclient.Space{Guid: spaceGuidOK, Name: spaceNameOK, OrgURL: "/v2/organizations/" + orgGuidOK}}
		space, _ := json.Marshal(spaceR)
		fmt.Fprint(w, string(space))
	})

	mux.HandleFunc("/v2/organizations/"+orgGuidOK, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		orgR := cfclient.OrgResource{Meta: cfclient.Meta{Guid: orgGuidOK}, Entity: cfclient.Org{Guid: orgGuidOK, Name: orgNameOK}}
		org, _ := json.Marshal(orgR)
		fmt.Fprint(w, string(org))
	})

	// Mock API for Error Case 1
	mux.HandleFunc("/v2/apps/"+appGuidErr1, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		appR := cfclient.AppResource{}
		appResource, _ := json.Marshal(appR)
		fmt.Fprint(w, string(appResource))

	})

	// Mock API for Error Case 2
	mux.HandleFunc("/v2/apps/"+appGuidErr2, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		appR := cfclient.AppResource{Meta: cfclient.Meta{Guid: appGuidErr2}, Entity: cfclient.App{Guid: appGuidErr2, Name: appNameErr2, SpaceURL: "/v2/spaces/" + spaceGuidErr2}}
		appResource, _ := json.Marshal(appR)
		fmt.Fprint(w, string(appResource))
	})

	// Mock API for Error Case 3
	mux.HandleFunc("/v2/apps/"+appGuidErr3, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		appR := cfclient.AppResource{Meta: cfclient.Meta{Guid: appGuidErr3}, Entity: cfclient.App{Guid: appGuidErr3, Name: appNameErr3, SpaceURL: "/v2/spaces/" + spaceGuidErr3}}
		appResource, _ := json.Marshal(appR)
		fmt.Fprint(w, string(appResource))
	})

	mux.HandleFunc("/v2/spaces/"+spaceGuidErr3, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		spaceR := cfclient.SpaceResource{Meta: cfclient.Meta{Guid: spaceGuidErr3}, Entity: cfclient.Space{Guid: spaceGuidErr3, Name: spaceNameErr3, OrgURL: "/v2/organizations/" + orgGuidErr3}}
		space, _ := json.Marshal(spaceR)
		fmt.Fprint(w, string(space))
	})

	tests := []struct {
		name            string
		appGUID         string
		wantErr         bool
		wantAppMetadata AppMetadata
	}{
		{"Get Metadata successfully", appGuidOK, false, AppMetadata{App: appNameOK, Space: spaceNameOK, Org: orgNameOK, AppGUID: appGuidOK, SpaceGUID: spaceGuidOK, OrgGUID: orgGuidOK}},
		{"Get App Metadata Error", appGuidErr1, true, AppMetadata{}},
		{"Get Space Metadata Error", appGuidErr2, true, AppMetadata{}},
		{"Get Org Metadata Error", appGuidErr3, true, AppMetadata{}},
	}

	for _, test := range tests {
		appMeta, err := cfClient.GetAppMetadata(test.appGUID)
		if !reflect.DeepEqual(appMeta, test.wantAppMetadata) || (err != nil) != test.wantErr {
			t.Fatalf("TestCFGetAppMetadata %s: expected %v, got %v, error = %v, wantErr %v", test.name, test.wantAppMetadata, appMeta, err, test.wantErr)
		}
	}
}

func TestCFGetRunningAppMetadata(t *testing.T) {
	teardown := setup()
	defer teardown()

	mux.HandleFunc("/v2/organizations", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		orgResp := cfclient.OrgResponse{Count: 1, Pages: 1, Resources: []cfclient.OrgResource{cfclient.OrgResource{Meta: cfclient.Meta{Guid: orgGuidOK}, Entity: cfclient.Org{Guid: orgGuidOK, Name: orgNameOK}}}}
		orgs, _ := json.Marshal(orgResp)
		fmt.Fprint(w, string(orgs))
	})

	// Tese case: get orgs only.
	allAppMeta, err := cfClient.GetRunningAppMetadata()
	if allAppMeta != nil || err == nil {
		t.Fatalf("TestCFGetRunningAppMetadata: expected nil, got %v, error = %v", allAppMeta, err)
	}

	mux.HandleFunc("/v2/spaces", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		spaceResp := cfclient.SpaceResponse{Count: 1, Pages: 1, Resources: []cfclient.SpaceResource{cfclient.SpaceResource{Meta: cfclient.Meta{Guid: spaceGuidOK}, Entity: cfclient.Space{Guid: spaceGuidOK, Name: spaceNameOK, OrgURL: "/v2/organizations/" + orgGuidOK, OrganizationGuid: orgGuidOK}}}}
		spaces, _ := json.Marshal(spaceResp)
		fmt.Fprint(w, string(spaces))
	})

	// Tese case: get orgs and spaces.
	allAppMeta, err = cfClient.GetRunningAppMetadata()
	if allAppMeta != nil || err == nil {
		t.Fatalf("TestCFGetRunningAppMetadata: expected nil, got %v, error = %v", allAppMeta, err)
	}

	mux.HandleFunc("/v2/apps", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		appResp := cfclient.AppResponse{Count: 2, Pages: 1, Resources: []cfclient.AppResource{cfclient.AppResource{Meta: cfclient.Meta{Guid: appGuidOK}, Entity: cfclient.App{Guid: appGuidOK, Name: appNameOK, SpaceURL: "/v2/spaces/" + spaceGuidOK, SpaceGuid: spaceGuidOK, State: "STARTED"}}, cfclient.AppResource{Meta: cfclient.Meta{Guid: appGuidNotStarted}, Entity: cfclient.App{Guid: appGuidNotStarted, Name: appNameNotStarted, SpaceURL: "/v2/spaces/" + spaceGuidOK, SpaceGuid: spaceGuidOK, State: "STOPPED"}}}}
		apps, _ := json.Marshal(appResp)
		fmt.Fprint(w, string(apps))
	})

	// Tese case: get orgs, spaces and apps(one app is started, the other is stopped).
	wantAppMetadata := []AppMetadata{AppMetadata{App: appNameOK, Space: spaceNameOK, Org: orgNameOK, AppGUID: appGuidOK, SpaceGUID: spaceGuidOK, OrgGUID: orgGuidOK}}
	allAppMeta, err = cfClient.GetRunningAppMetadata()
	if !reflect.DeepEqual(allAppMeta, wantAppMetadata) || err != nil {
		t.Fatalf("TestCFGetRunningAppMetadata: expected %v, got %v, error = %v", wantAppMetadata, allAppMeta, err)
	}
}
