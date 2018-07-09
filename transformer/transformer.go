package transformer

import (
	"encoding/binary"
	"fmt"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/pkg/errors"
	"github.com/rakutentech/cf-metrics-refinery/enricher"
)

type Envelope struct {
	Event  *events.Envelope
	Meta   enricher.AppMetadata
	Input  interface{}
	Output interface{}
}

var ErrEventDiscarded = errors.New("event discarded")

func (e *Envelope) AppGuid() string {
	switch e.Event.GetEventType() {
	case events.Envelope_HttpStartStop:
		return uuid2str(e.Event.GetHttpStartStop().GetApplicationId())

	case events.Envelope_LogMessage:
		return e.Event.GetLogMessage().GetAppId()

	case events.Envelope_ContainerMetric:
		return e.Event.GetContainerMetric().GetApplicationId()
	}

	return ""
}

func (e *Envelope) Enrich(E enricher.Enricher) error {
	appGUID := e.AppGuid()
	if appGUID == "" {
		return errors.New("envelope does not contain an app GUID")
	}

	md, err := E.GetAppMetadata(appGUID)
	if err != nil {
		return errors.Wrap(err, "getting app metadata for envelope")
	}

	e.Meta = md
	return nil
}

func uuid2str(uuid *events.UUID) string {
	if uuid == nil {
		return ""
	}
	var uuidBytes [16]byte
	binary.LittleEndian.PutUint64(uuidBytes[:8], uuid.GetLow())
	binary.LittleEndian.PutUint64(uuidBytes[8:], uuid.GetHigh())
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuidBytes[0:4], uuidBytes[4:6], uuidBytes[6:8], uuidBytes[8:10], uuidBytes[10:])
}
