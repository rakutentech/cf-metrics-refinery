package transformer

import (
	"encoding/json"
	"testing"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/rakutentech/cf-metrics-refinery/enricher"
)

func TestToInfluxDBPoint(t *testing.T) {
	tests := []struct {
		name         string
		msg          string
		appMeta      enricher.AppMetadata
		wantPointNil bool
		wantErr      bool
	}{
		{"APP log message", LogMsg, appMeta, false, false},
		{"RTR log message", RTRLogMsg, appMeta, false, false},
		{"Unknown log message EventDiscarded", UnknownLogMsg, appMeta, true, true},
		{"container metrics", containerMetrics, appMeta, false, false},
		{"httpStartStop", httpStartStop, appMeta, false, false},
		{"None metadata EventDiscarded", LogMsg, enricher.AppMetadata{}, true, true},
	}

	for _, test := range tests {
		var e events.Envelope
		if err := json.Unmarshal([]byte(test.msg), &e); err != nil {
			t.Fatal(err)
		}

		p, err := ToInfluxDBPoint(&Envelope{Event: &e, Meta: test.appMeta})
		if (err != nil) != test.wantErr {
			t.Fatalf("TestToInfluxDBPoint %s: error = %v, wantErr %v", test.name, err, test.wantErr)
		}
		if (p == nil) != test.wantPointNil {
			t.Fatalf("TestToInfluxDBPoint %s: point = %v, wantNil %v", test.name, p, test.wantPointNil)
		}
	}

}

func TestConvertLogMessage(t *testing.T) {
	tests := []struct {
		name   string
		logMsg string
	}{
		{"App log message", AppLogMsg},
		{"APP log message", LogMsg},
		{"RTR log message", RTRLogMsg},
	}

	for _, test := range tests {
		var e events.Envelope
		if err := json.Unmarshal([]byte(test.logMsg), &e); err != nil {
			t.Fatal(err)
		}

		point, err := convertLogMessage(e.GetLogMessage(), appMeta)
		if err != nil {
			t.Fatal(err)
		}

		if point.Name() != "log" {
			t.Fatalf("TestConvertLogMessage %s: expected %v got %v", test.name, "log", point.Name())
		}

		tags := []string{"app", "app_guid", "space", "space_guid", "org", "org_guid", "instance"}
		for _, key := range tags {
			if point.Tags()[key] != logPointTags[key] {
				t.Fatalf("TestConvertLogMessage %s: expected %v got %v, key %v", test.name, logPointTags, point.Tags(), key)
			}
		}

		fields := []string{"count", "size"}
		pointFields, _ := point.Fields()
		for _, key := range fields {
			if pointFields[key] != logPointFields[key] {
				t.Fatalf("TestConvertLogMessage %s: expected %v got %v %v %v %v", test.name, key, pointFields[key], logPointFields[key], logPointFields, pointFields)
			}
		}

		if point.UnixNano() != logPointUnixNano {
			t.Fatalf("TestConvertLogMessage %s: expected %v got %v", test.name, logPointUnixNano, point.UnixNano())
		}
	}
}

func TestConvertUnknownLogMessage(t *testing.T) {
	var e events.Envelope
	if err := json.Unmarshal([]byte(UnknownLogMsg), &e); err != nil {
		t.Fatal(err)
	}

	point, err := convertLogMessage(e.GetLogMessage(), appMeta)
	if err != ErrEventDiscarded || point != nil {
		t.Fatalf("TestConvertUnknownLogMessage expected %v got %v", ErrEventDiscarded, err)
	}
}

func TestConvertContainerMetric(t *testing.T) {
	var e events.Envelope
	if err := json.Unmarshal([]byte(containerMetrics), &e); err != nil {
		t.Fatal(err)
	}

	point, err := convertContainerMetric(e.GetContainerMetric(), e.GetTimestamp(), appMeta1)
	if err != nil {
		t.Fatal(err)
	}

	if point.Name() != "instance" {
		t.Fatalf("TestConvertContainerMetric expected %v got %v", "instance", point.Name())
	}

	tags := []string{"app", "app_guid", "space", "space_guid", "org", "org_guid", "instance"}
	for _, key := range tags {
		if point.Tags()[key] != containerMetricsPointTags[key] {
			t.Fatalf("TestConvertContainerMetric expected %v got %v", containerMetricsPointTags, point.Tags())
		}
	}

	pointFields, _ := point.Fields()
	fields := []string{"memory", "disk", "memory_quota", "disk_quota"}
	for _, key := range fields {
		if pointFields[key].(int64) != int64(containerMetricsPointFields[key].(int)) {
			t.Fatalf("TestConvertContainerMetric expected %v got %v", containerMetricsPointFields, pointFields)
		}
	}
	fields2 := []string{"cpu", "memory_pct", "disk_pct"}
	for _, key := range fields2 {
		if pointFields[key] != containerMetricsPointFields[key] {
			t.Fatalf("TestConvertContainerMetric expected %v got %v", containerMetricsPointFields, pointFields)
		}
	}

	if point.UnixNano() != containerMetricsPointUnixNano {
		t.Fatalf("TestConvertContainerMetric expected %v got %v", containerMetricsPointUnixNano, point.UnixNano())
	}
}

func TestConvertHttpStartStop(t *testing.T) {
	var e events.Envelope
	if err := json.Unmarshal([]byte(httpStartStop), &e); err != nil {
		t.Fatal(err)
	}

	point, err := convertHttpStartStop(e.GetHttpStartStop(), appMeta2)
	if err != nil {
		t.Fatal(err)
	}

	if point.Name() != "http_request" {
		t.Fatalf("TestConvertHttpStartStop expected %v got %v", "http_request", point.Name())
	}

	tags := []string{"app", "app_guid", "space", "space_guid", "org", "org_guid", "instance", "method", "status_code"}
	for _, key := range tags {
		if point.Tags()[key] != httpStartStopPointTags[key] {
			t.Fatalf("TestConvertHttpStartStop expected %v got %v", httpStartStopPointTags, point.Tags())
		}
	}

	pointFields, _ := point.Fields()
	fields1 := []string{"duration"}
	for _, key := range fields1 {
		if pointFields[key].(float64) != float64(httpStartStopPointFields[key].(int)) {
			t.Fatalf("TestConvertHttpStartStop expected %v got %v", httpStartStopPointFields, pointFields)
		}
	}
	fields2 := []string{"response_size", "count"}
	for _, key := range fields2 {
		if pointFields[key].(int64) != int64(httpStartStopPointFields[key].(int)) {
			t.Fatalf("TestConvertHttpStartStop expected %v got %v", httpStartStopPointFields, pointFields)
		}
	}
}
