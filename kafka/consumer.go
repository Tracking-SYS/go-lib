package kafka

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Tracking-SYS/go-lib/kafka/ccloud"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var mapMutex = sync.RWMutex{}

// ConsumeRecordValue represents the struct of the value in a Kafka message
type ConsumeRecordValue interface{}

func Start(consumerOuput chan []byte, topic string) {

	// Initialization
	mapMutex.Lock()
	configFile := ccloud.ParseArgs()
	conf := ccloud.ReadCCloudConfig(*configFile)
	mapMutex.Unlock()
	fmt.Printf("Kafka bootstrap servers: %v\n", conf[ccloud.BOOTSTRAP_SERVERS])
	// Create Consumer instance
	kafkaConfig := &kafka.ConfigMap{
		"session.timeout.ms":              10000,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		"default.topic.config":            kafka.ConfigMap{"auto.offset.reset": "earliest"},
	}

	for k, v := range conf {
		if v != "" {
			err := kafkaConfig.SetKey(k, v)
			if err != nil {
				fmt.Printf("Failed to set key: %s\n", err)
				os.Exit(1)
			}
		}
	}

	kafkaConfigJSON, err := json.Marshal(kafkaConfig)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}
	fmt.Printf("%v\n", string(kafkaConfigJSON))

	c, err := kafka.NewConsumer(kafkaConfig)
	if err != nil {
		fmt.Printf("Failed to create consumer: %s\n", err)
		os.Exit(1)
	}

	// Subscribe to topic
	err = c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		fmt.Printf("SubscribeTopics has error: %v\n", err)
	}

	fmt.Printf("Topic %v has been subscribed\n", topic)
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
		case ev := <-c.Events():
			switch e := ev.(type) {
			case kafka.AssignedPartitions:
				err = c.Assign(e.Partitions)
				if err != nil {
					fmt.Printf("Assign has error: %v\n", err)
				} else {
					fmt.Printf("AssignedPartitions: %v\n", e.Partitions)
				}
			case kafka.RevokedPartitions:
				err = c.Unassign()
				if err != nil {
					fmt.Printf("Unassign has error: %v\n", err)
				} else {
					fmt.Printf("RevokedPartitions: %v\n", e)
				}
			case *kafka.Message:
				fmt.Printf("Message on %s:\n%s\n",
					e.TopicPartition, string(e.Value))
				consumerOuput <- e.Value
			case kafka.PartitionEOF:
				fmt.Printf("Reached %v\n", e)
			case kafka.Error:
				// Errors should generally be considered as informational, the client will try to automatically recover
				fmt.Printf("Kafka Error: %v\n", e)
			}
		}
	}

	fmt.Printf("Closing consumer\n")
	close(consumerOuput)
	c.Close()
}
