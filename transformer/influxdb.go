package transformer

import (
	"fmt"
	"strings"
	"time"

	"github.com/cloudfoundry/sonde-go/events"
	influxdb "github.com/influxdata/influxdb/client/v2"
	"github.com/rakutentech/cf-metrics-refinery/enricher"
)

func ToInfluxDBPoint(event *Envelope) (*influxdb.Point, error) {
	if event.Meta.App == "" {
		return nil, ErrEventDiscarded
	}

	switch event.Event.GetEventType() {
	default:
		return nil, ErrEventDiscarded

	case events.Envelope_HttpStartStop:
		return convertHttpStartStop(event.Event.GetHttpStartStop(), event.Meta)

	case events.Envelope_LogMessage:
		return convertLogMessage(event.Event.GetLogMessage(), event.Meta)

	case events.Envelope_ContainerMetric:
		return convertContainerMetric(event.Event.GetContainerMetric(), event.Event.GetTimestamp(), event.Meta)
	}
}

func convertHttpStartStop(e *events.HttpStartStop, meta enricher.AppMetadata) (*influxdb.Point, error) {
	start := time.Unix(0, e.GetStartTimestamp())
	stop := time.Unix(0, e.GetStopTimestamp())

	return influxdb.NewPoint(
		"http_request", // metric name
		map[string]string{ // tags
			"app":         meta.App,
			"app_guid":    meta.AppGUID,
			"space":       meta.Space,
			"space_guid":  meta.SpaceGUID,
			"org":         meta.Org,
			"org_guid":    meta.OrgGUID,
			"instance":    fmt.Sprint(e.GetInstanceIndex()),
			"method":      e.GetMethod().String(),
			"status_code": fmt.Sprint(e.GetStatusCode()),
			// "instance_guid": e.GetInstanceId(),
		},
		map[string]interface{}{ // values
			"count":         1, // Not needed but for convenience and furthur usage.
			"duration":      stop.Sub(start).Seconds(),
			"response_size": e.GetContentLength(),
		},
		start, // timestamp
	)
}

func convertLogMessage(e *events.LogMessage, meta enricher.AppMetadata) (*influxdb.Point, error) {
	if strings.HasPrefix(e.GetSourceType(), "APP") || strings.HasPrefix(e.GetSourceType(), "App") {
		return convertAppLogMessage(e, meta)
	} else if strings.HasPrefix(e.GetSourceType(), "RTR") {
		return convertRtrLogMessage(e, meta)
	} else {
		return nil, ErrEventDiscarded
	}
}

func convertAppLogMessage(e *events.LogMessage, meta enricher.AppMetadata) (*influxdb.Point, error) {
	return influxdb.NewPoint(
		"log",
		map[string]string{
			"app":        meta.App,
			"app_guid":   meta.AppGUID,
			"space":      meta.Space,
			"space_guid": meta.SpaceGUID,
			"org":        meta.Org,
			"org_guid":   meta.OrgGUID,
			"instance":   e.GetSourceInstance(),
			"type":       e.GetMessageType().String(),
			// "instance_guid": e.???,
		},
		map[string]interface{}{
			"count": 1, // Not needed but included for convenience.
			"size":  len(e.GetMessage()),
		},
		time.Unix(0, e.GetTimestamp()),
	)
}

func convertRtrLogMessage(e *events.LogMessage, meta enricher.AppMetadata) (*influxdb.Point, error) {
	return influxdb.NewPoint(
		"log",
		map[string]string{
			"app":        meta.App,
			"app_guid":   meta.AppGUID,
			"space":      meta.Space,
			"space_guid": meta.SpaceGUID,
			"org":        meta.Org,
			"org_guid":   meta.OrgGUID,
			"instance":   e.GetSourceInstance(),
			"type":       "RTR",
		},
		map[string]interface{}{
			"count": 1, // Not needed but included for convenience.
			"size":  len(e.GetMessage()),
		},
		time.Unix(0, e.GetTimestamp()),
	)
}

func convertContainerMetric(e *events.ContainerMetric, ts int64, meta enricher.AppMetadata) (*influxdb.Point, error) {
	return influxdb.NewPoint(
		"instance",
		map[string]string{
			"app":        meta.App,
			"app_guid":   meta.AppGUID,
			"space":      meta.Space,
			"space_guid": meta.SpaceGUID,
			"org":        meta.Org,
			"org_guid":   meta.OrgGUID,
			"instance":   fmt.Sprint(e.GetInstanceIndex()),
			// "instance_guid": e.???,
		},
		map[string]interface{}{
			"cpu":          e.GetCpuPercentage(),
			"memory":       int64(e.GetMemoryBytes()),
			"disk":         int64(e.GetDiskBytes()),
			"memory_quota": int64(e.GetMemoryBytesQuota()),
			"disk_quota":   int64(e.GetDiskBytesQuota()),
			"memory_pct":   float64(e.GetMemoryBytes()) / float64(e.GetMemoryBytesQuota()),
			"disk_pct":     float64(e.GetDiskBytes()) / float64(e.GetDiskBytesQuota()),
		},
		time.Unix(0, ts),
	)
}
