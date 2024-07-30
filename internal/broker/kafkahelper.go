package broker

import (
	"fmt"
	"time"

	"brandlovrs.exporter/internal/config"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func ConsumeFromTopic(topicName string) ([]byte, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": config.ReadParameter("KAFKA_BROKERS"),
		"group.id":          config.ReadParameter("KAFKA_GROUP_ID"),
		"auto.offset.reset": config.ReadParameter("KAFKA_AUTO_OFFSET_RESET"),
	})
	if err != nil {
		panic(err)
	}

	err = c.SubscribeTopics([]string{topicName, "^aRegex.*[Tt]opic"}, nil)

	// A signal handler or similar could be used to set this to false to break the loop.
	run := true

	for run {
		msg, err := c.ReadMessage(time.Second)
		if err == nil {
			// TODO: Remove this after the initial tests...
			fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))

			c.Close()
			return msg.Value, err
		} else if !err.(kafka.Error).IsTimeout() {
			if err != nil || err == nil {
				continue
			}
		}
	}

	return nil, err
}
