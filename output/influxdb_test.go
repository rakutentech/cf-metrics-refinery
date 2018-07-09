package output

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/rakutentech/cf-metrics-refinery/enricher"
	"github.com/rakutentech/cf-metrics-refinery/transformer"
)

func TestInfluxDBLogMessage(t *testing.T) {
	var tests = []struct {
		src     string
		appMeta enricher.AppMetadata
		res     string
	}{
		{AppOutLogMsg, appMeta, "log,app=app,app_guid=00000000-0000-0000-0000-000000000000,instance=0,org=org,org_guid=20000000-0000-0000-0000-000000000000,space=space,space_guid=10000000-0000-0000-0000-000000000000,type=OUT count=1i,size=12i 123456789012345000\n"},
		{AppErrLogMsg, appMeta, "log,app=app,app_guid=00000000-0000-0000-0000-000000000000,instance=1,org=org,org_guid=20000000-0000-0000-0000-000000000000,space=space,space_guid=10000000-0000-0000-0000-000000000000,type=ERR count=1i,size=16i 123456789000000000\n"},
		{RtrLogMsg, appMeta, "log,app=app,app_guid=00000000-0000-0000-0000-000000000000,instance=2,org=org,org_guid=20000000-0000-0000-0000-000000000000,space=space,space_guid=10000000-0000-0000-0000-000000000000,type=RTR count=1i,size=12i 123456789012345000\n"},
		{nonAppGuidLogMsg, noneAppMeta, ""},
	}
	for _, test := range tests {
		done := make(chan []byte, 1)

		srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			defer close(done)
			if res, err := ioutil.ReadAll(r.Body); err != nil {
				panic(err)
			} else {
				done <- res
			}
		}))
		defer srv.Close()

		o, err := NewInfluxDB(ConfigInfluxDB{
			Addr:     srv.URL,
			Database: "test",
		})
		if err != nil {
			t.Fatal(err)
		}

		var e events.Envelope
		if err := json.Unmarshal([]byte(test.src), &e); err != nil {
			t.Fatal(err)
		}

		env := &transformer.Envelope{Meta: test.appMeta, Event: &e}

		if err := o.Write(env); err != nil {
			t.Fatal(err)
		}

		res := <-done
		if bytes.Compare(res, []byte(test.res)) != 0 {
			t.Fatalf("expected %q got %q", test.res, res)
		}
	}
}

func BenchmarkInfluxDB(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	defer srv.Close()

	o, err := NewInfluxDB(ConfigInfluxDB{
		Addr:     srv.URL,
		Database: "test",
	})
	if err != nil {
		b.Fatal(err)
	}

	var e events.Envelope
	if err := json.Unmarshal([]byte(AppOutLogMsg), &e); err != nil {
		b.Fatal(err)
	}

	env := &transformer.Envelope{Meta: appMeta, Event: &e}
	envs := []*transformer.Envelope{}

	for i := 0; i < b.N; i++ {
		envs = append(envs, env)
	}

	b.ResetTimer()
	if err := o.Write(envs...); err != nil {
		b.Fatal(err)
	}
}
