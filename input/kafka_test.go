package input

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/cloudfoundry/sonde-go/events"
)

func TestProcess(t *testing.T) {
	msgFailed2Unmarshal := &sarama.ConsumerMessage{
		Topic:     "cf-app-log-test-fail",
		Partition: 10,
		Offset:    119373397,
		Value:     []byte("This message is for testing"),
	}
	offsetsFail := make(map[string]map[int32]int64)
	offsetsFail[msgFailed2Unmarshal.Topic] = make(map[int32]int64)

	origin := "rep"
	var eventType events.Envelope_EventType = 5
	timestamp := int64(1527158917416335302)
	deployment := "lab-diego-openstack"
	job := "cell_dev"
	index := "23acb41b-6979-4a03-ba66-5e0ff08908f7"
	ip := "100.73.61.130"
	message := "This log message is for testing."
	var messageType events.LogMessage_MessageType = 1
	logTimestamp := int64(1527158917416336022)
	appId := "fc0f097f-cd4f-4478-9f82-c99462611f4c"
	sourceType := "APP/PROC/WEB"
	sourceInstance := "0"

	logMessage := events.LogMessage{
		Message:        []byte(message),
		MessageType:    &messageType,
		Timestamp:      &logTimestamp,
		AppId:          &appId,
		SourceType:     &sourceType,
		SourceInstance: &sourceInstance,
	}
	logMsg := events.Envelope{
		Origin:     &origin,
		EventType:  &eventType,
		Timestamp:  &timestamp,
		Deployment: &deployment,
		Job:        &job,
		Index:      &index,
		Ip:         &ip,
		LogMessage: &logMessage,
	}

	log, _ := json.Marshal(logMsg)
	logM := &sarama.ConsumerMessage{
		Topic:     "cf-app-log-test",
		Partition: 10,
		Offset:    119373397,
		Value:     log,
	}

	offsetsPartition := make(map[int32]int64)
	offsetsPartition[logM.Partition] = logM.Offset
	offsetsTopic := make(map[string]map[int32]int64)
	offsetsTopic[logM.Topic] = offsetsPartition

	tests := []struct {
		name        string
		msg         *sarama.ConsumerMessage
		wantErr     bool
		wantOffsets map[string]map[int32]int64
	}{
		{"cf-app-log msg", logM, false, offsetsTopic},
		{"msg failed to unmarshal", msgFailed2Unmarshal, true, offsetsFail},
	}

	for _, test := range tests {
		c := &KafkaConsumer{Offsets: make(map[string]map[int32]int64)}
		_, err := c.Process(test.msg)
		if (err != nil) != test.wantErr {
			t.Fatalf("TestProcess, %s: error = %v, wantErr %v", test.name, err, test.wantErr)
		}
		if !reflect.DeepEqual(c.Offsets, test.wantOffsets) {
			t.Fatalf("TestProcess, %s: expected Offsets %v, got %v", test.name, c.Offsets, test.wantOffsets)
		}
	}

}
