package input

import (
	"encoding/json"
	"log"
	"time"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	"github.com/rakutentech/cf-metrics-refinery/transformer"
	"github.com/wvanbergen/kafka/consumergroup"
	"github.com/wvanbergen/kazoo-go"
)

type ConfigKafka struct {
	Zookeepers        string        `required:"true" desc:"Zookeeper nodes for offset storage"`                                                              // CFMR_KAFKA_ZOOKEEPERS
	Topics            []string      `required:"true" desc:"Topics to read events from"`                                                                      // CFMR_KAFKA_TOPICS
	ConsumerGroup     string        `required:"true" desc:"Name of the Kafka consumer group"`                                                                // CFMR_KAFKA_CONSUMERGROUP
	ProcessingTimeout time.Duration `default:"1m" desc:"Time to wait for all the offsets for a partition to be processed after stopping to consume from it"` // CFMR_KAFKA_PROCESSINGTIMEOUT
	OffsetNewest      bool          `default:"false" desc:"If true start from the newest message in Kafka in case the offset in zookeeper does not exist"`   // CFMR_KAFKA_OFFSETNEWEST
}

type KafkaConsumer struct {
	// FIXME: these should be private
	CG      *consumergroup.ConsumerGroup
	Offsets map[string]map[int32]int64
}

// Close Kafka consumer group
func (c *KafkaConsumer) Close() error {
	return errors.Wrap(c.CG.Close(), "closing kafka consumer group")
}

// Create Kafka consumer group
func NewKafkaConsumer(i *ConfigKafka) (*KafkaConsumer, error) {
	config := consumergroup.NewConfig()
	if i.OffsetNewest {
		config.Offsets.Initial = sarama.OffsetNewest
	}
	config.Offsets.ProcessingTimeout = i.ProcessingTimeout

	zkNodes, zkChroot := kazoo.ParseConnectionString(i.Zookeepers)
	config.Zookeeper.Chroot = zkChroot

	consumer, consumerErr := consumergroup.JoinConsumerGroup(i.ConsumerGroup, i.Topics, zkNodes, config)
	if consumerErr != nil {
		return nil, errors.Wrap(consumerErr, "Failed to join Kafka consumer group")
	}

	return &KafkaConsumer{
		CG:      consumer,
		Offsets: make(map[string]map[int32]int64),
	}, nil
}

// Read message from Kafka
func (c *KafkaConsumer) Read() (*transformer.Envelope, error) {
	// Consume from Kafka
	message, ok := <-c.CG.Messages()
	if !ok {
		return nil, errors.New("Failed to consume data from Kafka")
	}
	return c.Process(message)
}

func (c *KafkaConsumer) Process(message *sarama.ConsumerMessage) (*transformer.Envelope, error) {
	// FIXME: do we really need all these checks for the correct offsets?
	t, found := c.Offsets[message.Topic]
	if !found {
		t = make(map[int32]int64)
		c.Offsets[message.Topic] = t
	}

	// make sure we get the message with the offset we're expecting
	o, found := t[message.Partition]
	if found && o+1 != message.Offset {
		log.Printf(
			"Unexpected offset on %s:%d. Expected %d, found %d, diff %d",
			message.Topic, message.Partition, o+1, message.Offset, message.Offset-(o+1))
	}

	envelope := &transformer.Envelope{}
	err := json.Unmarshal(message.Value, &envelope.Event)
	if err != nil {
		// FIXME: is this correct? we need to make sure that a single broken event does
		// not stop processing completely
		return nil, errors.Wrap(err, "failed to unmarshal event")
	}

	t[message.Partition] = message.Offset
	envelope.Input = message

	return envelope, nil
}
