package transformer

import (
	"github.com/rakutentech/cf-metrics-refinery/enricher"
)

const LogMsg = `{
	"origin": "rep",
	"eventType": 5,
	"timestamp": 123456789012345678,
	"job": "cell",
	"index": "0",
	"ip": "192.168.0.50",
	"logMessage": {
		"message": "aGVsbG8gd29ybGQK",
		"message_type": 1,
		"timestamp": 123456789012345000,
		"app_id": "00000000-0000-0000-0000-000000000000",
		"source_type": "APP",
		"source_instance": "1"
	}
}`

const AppLogMsg = `{
	"origin": "rep",
	"eventType": 5,
	"timestamp": 123456789012345678,
	"job": "cell",
	"index": "0",
	"ip": "192.168.0.50",
	"logMessage": {
		"message": "aGVsbG8gd29ybGQK",
		"message_type": 1,
		"timestamp": 123456789012345000,
		"app_id": "00000000-0000-0000-0000-000000000000",
		"source_type": "App",
		"source_instance": "1"
	}
}`

const RTRLogMsg = `{
	"origin": "gorouter",
	"eventType": 5,
	"timestamp": 123456789012345678,
	"job": "router",
	"index": "c9ba631e-cbb7-42c7-b4d4-74adaf140685",
	"ip": "192.168.0.50",
	"logMessage": {
		"message": "aGVsbG8gd29ybGQK",
		"message_type": 1,
		"timestamp": 123456789012345000,
		"app_id": "00000000-0000-0000-0000-000000000003",
		"source_type": "RTR",
		"source_instance": "1"
	}
}`

const UnknownLogMsg = `{
	"origin": "Unknown",
	"eventType": 5,
	"timestamp": 123456789012345678,
	"job": "Unknown",
	"index": "c9ba631e-cbb7-42c7-b4d4-74adaf140685",
	"ip": "192.168.0.50",
	"logMessage": {
		"message": "aGVsbG8gd29ybGQK",
		"message_type": 1,
		"timestamp": 123456789012345000,
		"app_id": "00000000-0000-0000-0000-000000000003",
		"source_type": "Unknown",
		"source_instance": "1"
	}
}`

const NonAppGuidLogMsg = `{
	"origin": "rep",
	"eventType": 5,
	"timestamp": 123456789012345678,
	"job": "cell",
	"index": "0",
	"ip": "192.168.0.50",
	"logMessage": {
		"message": "aGVsbG8gd29ybGQK",
		"message_type": 0,
		"timestamp": 123456789012345000,
		"app_id": "",
		"source_type": "APP",
		"source_instance": "1"
	}
}`

var appMeta = enricher.AppMetadata{
	App:       "app",
	AppGUID:   "00000000-0000-0000-0000-000000000000",
	Space:     "space",
	SpaceGUID: "10000000-0000-0000-0000-000000000000",
	Org:       "org",
	OrgGUID:   "20000000-0000-0000-0000-000000000000",
}

var logPointTags = map[string]string{
	"app":        "app",
	"app_guid":   "00000000-0000-0000-0000-000000000000",
	"instance":   "1",
	"org":        "org",
	"org_guid":   "20000000-0000-0000-0000-000000000000",
	"space":      "space",
	"space_guid": "10000000-0000-0000-0000-000000000000",
	"type":       "0",
}

var logPointFields = map[string]int64{
	"count": 1,
	"size":  12,
}

const logPointUnixNano = 123456789012345000

const containerMetrics = `{
	"origin": "rep",
	"eventType": 9,
	"timestamp": 123456789012345678,
	"job": "cell",
	"index": "0",
	"ip": "192.168.0.50",
	"containerMetric": {
		"applicationId": "00000000-0000-0000-0000-000000000001",
		"instanceIndex": 1,
		"cpuPercentage": 1.212661987393707,
		"memoryBytes": 54050816,
		"diskBytes": 109563904,
		"memoryBytesQuota": 67108864,
		"diskBytesQuota": 1073741824
	}
}`

var appMeta1 = enricher.AppMetadata{
	App:       "app1",
	AppGUID:   "00000000-0000-0000-0000-000000000001",
	Space:     "space1",
	SpaceGUID: "10000000-0000-0000-0000-000000000001",
	Org:       "org1",
	OrgGUID:   "20000000-0000-0000-0000-000000000001",
}

var containerMetricsPointTags = map[string]string{
	"app":        "app1",
	"app_guid":   "00000000-0000-0000-0000-000000000001",
	"instance":   "1",
	"org":        "org1",
	"org_guid":   "20000000-0000-0000-0000-000000000001",
	"space":      "space1",
	"space_guid": "10000000-0000-0000-0000-000000000001",
}

var containerMetricsPointFields = map[string]interface{}{
	"cpu":          1.212661987393707,
	"memory":       54050816,
	"disk":         109563904,
	"memory_quota": 67108864,
	"disk_quota":   1073741824,
	"memory_pct":   0.805419921875,
	"disk_pct":     0.10203933715820312,
}

const containerMetricsPointUnixNano = 123456789012345678

const httpStartStop = `{
	"origin": "gorouter",
	"eventType": 4,
	"timestamp": 123456789012345678,
	"job": "router",
	"index": "1",
	"ip": "192.168.0.50",
	"HttpStartStop": {
		"startTimestamp": 1524923912949154418,
		"stopTimestamp": 1524923912949154418,
		"requestId": {"low": 18034462508262158772, "high": 1299234503289247342},
		"peerType": 1,
		"method": 3,
		"uri": "http://registry.dev.lab-jpe2.rpaas.net/eureka/apps/BACKEND/6a243643-3d2d-4f44-5933-fa6f",
		"remoteAddress": "127.0.0.1:14130",
		"userAgent": "Java-EurekaClient/v1.4.11",
		"statusCode": 200,
		"contentLength": 0,
		"applicationId": {"low": 14285923797169022654, "high": 6940295952872734603},
		"instanceIndex": 1,
		"instanceId": "d06a0894-aea7-4b88-4860-08fe",
		"forwarded": ["100.73.61.130", "127.0.0.1"]
	}
}`

var appMeta2 = enricher.AppMetadata{
	App:       "app2",
	AppGUID:   "be268fe2-00cc-41c6-8b7f-0fdb65e25060",
	Space:     "space2",
	SpaceGUID: "10000000-0000-0000-0000-000000000002",
	Org:       "org2",
	OrgGUID:   "20000000-0000-0000-0000-000000000002",
}

var httpStartStopPointTags = map[string]string{
	"app":         "app2",
	"app_guid":    "be268fe2-00cc-41c6-8b7f-0fdb65e25060",
	"instance":    "1",
	"org":         "org2",
	"org_guid":    "20000000-0000-0000-0000-000000000002",
	"space":       "space2",
	"space_guid":  "10000000-0000-0000-0000-000000000002",
	"method":      "PUT",
	"status_code": "200",
}

var httpStartStopPointFields = map[string]interface{}{
	"count":         1,
	"duration":      0,
	"response_size": 0,
}
