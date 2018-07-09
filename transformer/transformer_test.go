package transformer

import (
	"encoding/json"
	"testing"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/pkg/errors"
	"github.com/rakutentech/cf-metrics-refinery/enricher"
)

func TestUuid2Str(t *testing.T) {
	tests := []struct {
		high     uint64
		low      uint64
		expected string
	}{
		{0x0001020304050607, 0x08090A0B0C0D0E0F, "0f0e0d0c-0b0a-0908-0706-050403020100"},
		{0x0011223344556677, 0x8899AABBCCDDEEFF, "ffeeddcc-bbaa-9988-7766-554433221100"},
		{0x0000000000000000, 0x0000000000000000, "00000000-0000-0000-0000-000000000000"},
	}

	for _, test := range tests {
		uuid := &events.UUID{
			High: &test.high,
			Low:  &test.low,
		}
		res := uuid2str(uuid)
		if res != test.expected {
			t.Fatalf("expected %q got %q", test.expected, res)
		}
	}
}

func TestUuid2StrNil(t *testing.T) {
	if res := uuid2str(nil); res != "" {
		t.Errorf("expected %q got %q", "", res)
	}
}

func TestAppGuid(t *testing.T) {
	tests := []struct {
		name        string
		msg         string
		wantAppGuid string
	}{
		{"App log message", LogMsg, "00000000-0000-0000-0000-000000000000"},
		{"RTR log message", RTRLogMsg, "00000000-0000-0000-0000-000000000003"},
		{"container metrics", containerMetrics, "00000000-0000-0000-0000-000000000001"},
		{"httpStartStop", httpStartStop, "be268fe2-00cc-41c6-8b7f-0fdb65e25060"},
	}

	for _, test := range tests {
		var e events.Envelope
		if err := json.Unmarshal([]byte(test.msg), &e); err != nil {
			t.Fatal(err)
		}

		envelope := &Envelope{Event: &e, Meta: appMeta}
		appGuid := envelope.AppGuid()
		if appGuid != test.wantAppGuid {
			t.Fatalf("%s: appGuid = %v, wantAppGuid %v", test.name, appGuid, test.wantAppGuid)
		}
	}

}

type me struct {
	m map[string]enricher.AppMetadata
}

func (m *me) GetAppMetadata(appGUID string) (enricher.AppMetadata, error) {
	if md, ok := m.m[appGUID]; ok {
		return md, nil
	} else {
		return enricher.AppMetadata{}, errors.New("no such app")
	}
}

func mockEnricher(appGUID ...string) enricher.Enricher {
	m := make(map[string]enricher.AppMetadata, len(appGUID))
	for _, a := range appGUID {
		m[a] = mockData(a)
	}
	return &me{m}
}

func mockData(a string) enricher.AppMetadata {
	return enricher.AppMetadata{
		App:       "app" + a,
		AppGUID:   a,
		Space:     "space" + a,
		SpaceGUID: a,
		Org:       "org" + a,
		OrgGUID:   a,
	}
}

func TestEnrich(t *testing.T) {
	enricher := mockEnricher("00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000001", "be268fe2-00cc-41c6-8b7f-0fdb65e25060")
	tests := []struct {
		name       string
		msg        string
		wantErr    bool
		wantErrStr string
	}{
		{"App log message", LogMsg, false, ""},
		{"Container metrics", containerMetrics, false, ""},
		{"HttpStartStop", httpStartStop, false, ""},
		{"Empty appGuid log message", NonAppGuidLogMsg, true, "envelope does not contain an app GUID"},
		{"No metedata log message", RTRLogMsg, true, "getting app metadata for envelope: no such app"},
	}

	for _, test := range tests {
		var e events.Envelope
		if err := json.Unmarshal([]byte(test.msg), &e); err != nil {
			t.Fatal(err)
		}

		envelope := &Envelope{Event: &e, Meta: appMeta}
		err := envelope.Enrich(enricher)
		//if test.wantErr != err {
		if (err != nil) != test.wantErr || (err != nil && err.Error() != test.wantErrStr) {
			t.Fatalf("Transformer.Enrich() error = %v, wantErr %v", err, test.wantErrStr)
		}
	}

}
