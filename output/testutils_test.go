package output

import (
	"github.com/rakutentech/cf-metrics-refinery/enricher"
)

// this should probably be moved to input tests
const LogMsg = `{
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
		"app_id": "00000000-0000-0000-0000-000000000000",
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

const logPoint = "log,app=app,app_guid=00000000-0000-0000-0000-000000000000,instance=1,org=org,org_guid=20000000-0000-0000-0000-000000000000,space=space,space_guid=10000000-0000-0000-0000-000000000000,type=0 count=1i,size=12i 123456789012345000\n"

const nonAppGuidLogMsg = `{
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

var noneAppMeta = enricher.AppMetadata{}

const noneLogPoint = ""

const httpStartStopPoint = "http_request,app=app2,app_guid=00000000-0000-0000-0000-000000000002,instance=1,org=org2,org_guid=20000000-0000-0000-0000-000000000002,space=space2,space_guid=10000000-0000-0000-0000-000000000002,method=PUT,status_code=200 count=1i,duration=(1524923912951886498-1524923912949154418)/1000000000.0,response_size=0i 123456789012345002\n"

const AppOutLogMsg = `{
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
		"source_instance": "0"
	}
}`
const AppErrLogMsg = `{
	"origin": "rep",
	"eventType": 5,
	"timestamp": 123456789012345678,
	"job": "cell",
	"index": "0",
	"ip": "192.168.0.50",
	"logMessage": {
		"message": "eWFkZGF5YWRkYXlhZGRhCg==",
		"message_type": 2,
		"timestamp": 123456789000000000,
		"app_id": "00000000-0000-0000-0000-000000000000",
		"source_type": "APP",
		"source_instance": "1"
	}
}`
const RtrLogMsg = `{
	"origin": "gorouter",
	"eventType": 5,
	"timestamp": 123456789012345678,
	"job": "router",
	"index": "7",
	"ip": "192.168.1.50",
	"logMessage": {
		"message": "aGVsbG8gd29ybGQK",
		"message_type": 0,
		"timestamp": 123456789012345000,
		"app_id": "00000000-0000-0000-0000-000000000000",
		"source_type": "RTR",
		"source_instance": "2"
	}
}`
