package debug

import (
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestIndex(t *testing.T) {
	s, _ := NewServer(&ConfigServer{Port: "8080"}, &Stats{}, log.New(os.Stdout, "TEST", log.LstdFlags))
	srv := s.Start()
	defer func() { srv.Shutdown(nil) }()
	URL := "http://127.0.0.1:8080"

	//FIXME: need to Wait for the http server to start fully.
	time.Sleep(1 * time.Second)

	newreq := func(method, url string, body io.Reader) *http.Request {
		r, err := http.NewRequest(method, url, body)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}

	tests := []struct {
		name string
		r    *http.Request
	}{
		{name: "get /", r: newreq("GET", URL+"/", nil)},
		{name: "post /", r: newreq("POST", URL+"/", nil)},
		{name: "get /stats/runtime", r: newreq("GET", URL+"/stats/runtime", nil)},
		{name: "post /stats/runtime", r: newreq("POST", URL+"/stats/runtime", nil)},
		{name: "get /debug/pprof/", r: newreq("GET", URL+"/debug/pprof/", nil)},
		{name: "post /debug/pprof/", r: newreq("POST", URL+"/debug/pprof/", nil)},
		{name: "get /stats/app", r: newreq("GET", URL+"/stats/app", nil)},
		{name: "post /stats/app", r: newreq("POST", URL+"/stats/app", nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.DefaultClient.Do(tt.r)
			defer resp.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
