package kafka

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/lk153/go-lib/kafka/ccloud"
)

// ConsumeRecordValue represents the struct of the value in a Kafka message
type ConsumeRecordValue interface{}

const (
	METADATA_BROKER_LIST = "metadata.broker.list"
	BOOTSTRAP_SERVERS    = "bootstrap.servers"
	SASL_MECHANISMS      = "sasl.mechanisms"
	SECURITY_PROTOCOL    = "security.protocol"
	SASL_USERNAME        = "sasl.username"
	SASL_PASSWORD        = "sasl.password"
)

func Start() {
	// Initialization
	configFile, topic := ccloud.ParseArgs()
	conf := ccloud.ReadCCloudConfig(*configFile)

	fmt.Print(conf[BOOTSTRAP_SERVERS])
	// Create Consumer instance
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		METADATA_BROKER_LIST:              conf[METADATA_BROKER_LIST],
		BOOTSTRAP_SERVERS:                 conf[BOOTSTRAP_SERVERS],
		SASL_MECHANISMS:                   conf[SASL_MECHANISMS],
		SECURITY_PROTOCOL:                 conf[SECURITY_PROTOCOL],
		SASL_USERNAME:                     conf[SASL_USERNAME],
		SASL_PASSWORD:                     conf[SASL_PASSWORD],
		"group.id":                        "cloudkarafka-example",
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		"default.topic.config": kafka.ConfigMap{
			"auto.offset.reset": "earliest",
		},
	})
	if err != nil {
		fmt.Printf("Failed to create consumer: %s", err)
		os.Exit(1)
	}

	// Subscribe to topic
	err = c.SubscribeTopics([]string{*topic}, nil)
	if err != nil {
		fmt.Printf("SubscribeTopics has error: %v\n", err)
	}

	// Set up a channel for handling Ctrl-C, etc
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Process messages
	run := true
	for run {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			msg, err := c.ReadMessage(1000 * time.Millisecond)
			if err != nil {
				fmt.Printf("Failed to read message: %v", err)
				continue
			}
			recordKey := string(msg.Key)
			recordValue := msg.Value
			var data ConsumeRecordValue
			err = json.Unmarshal(recordValue, &data)
			if err != nil {
				fmt.Printf("Failed to decode JSON at offset %d: %v", msg.TopicPartition.Offset, err)
				continue
			}
			fmt.Printf("Consumed record with key %s and value %s\n", recordKey, recordValue)
		}
	}

	fmt.Printf("Closing consumer\n")
	c.Close()
}
