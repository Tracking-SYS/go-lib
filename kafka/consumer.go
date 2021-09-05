package kafka

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Tracking-SYS/go-lib/kafka/ccloud"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	confluentKafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

var mapMutex = sync.RWMutex{}

// ConsumeRecordValue represents the struct of the value in a Kafka message
type ConsumeRecordValue interface{}

//Consumer ...
type Consumer struct {
	ConfigFile *string
	Conf       map[string]string
	Consumer   *confluentKafka.Consumer
}

//InitConfig ...
func (kc *Consumer) InitConfig() error {
	if *kc.ConfigFile == "" {
		fmt.Printf("empty configFile\n")
		return fmt.Errorf("empty configFile")
	}

	mapMutex.Lock()
	kc.Conf = ccloud.ReadCCloudConfig(*kc.ConfigFile)
	mapMutex.Unlock()
	return nil
}

//CreateConsumerInstance ...
func (kc *Consumer) CreateConsumerInstance() error {
	// Initialization
	fmt.Printf("Kafka bootstrap servers: %v\n", kc.Conf[ccloud.BOOTSTRAP_SERVERS])
	// Create Consumer instance
	kafkaConfig := &kafka.ConfigMap{
		"session.timeout.ms":              10000,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		"default.topic.config":            kafka.ConfigMap{"auto.offset.reset": "earliest"},
	}

	//Set custom configuration from yaml
	for k, v := range kc.Conf {
		if v != "" {
			err := kafkaConfig.SetKey(k, v)
			if err != nil {
				fmt.Printf("Failed to set key: %s\n", err)
				return err
			}
		}
	}

	//Create consumer instance
	c, err := kafka.NewConsumer(kafkaConfig)
	if err != nil {
		fmt.Printf("Failed to create consumer: %s\n", err)
		return err
	}

	kc.Consumer = c
	return nil
}

//Start ...
func (kc *Consumer) Start(consumerOuput map[string]chan []byte, topics []string) {
	// Subscribe to topic
	err := kc.Consumer.SubscribeTopics(topics, nil)
	if err != nil {
		fmt.Printf("SubscribeTopics %s has error: %v\n", topics, err)
	}

	fmt.Printf("Topics %v has been subscribed\n", topics)
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
		case ev := <-kc.Consumer.Events():
			switch e := ev.(type) {
			case kafka.AssignedPartitions:
				err = kc.Consumer.Assign(e.Partitions)
				if err != nil {
					fmt.Printf("Assign has error: %v\n", err)
				} else {
					fmt.Printf("AssignedPartitions: %v\n", e.Partitions)
				}
			case kafka.RevokedPartitions:
				err = kc.Consumer.Unassign()
				if err != nil {
					fmt.Printf("Unassign has error: %v\n", err)
				} else {
					fmt.Printf("RevokedPartitions: %v\n", e)
				}
			case *kafka.Message:
				fmt.Printf("Message on %s:\n%s\n",
					e.TopicPartition, string(e.Value))
				consumerOuput[*e.TopicPartition.Topic] <- e.Value
			case kafka.PartitionEOF:
				fmt.Printf("Reached %v\n", e)
			case kafka.Error:
				// Errors should generally be considered as informational, the client will try to automatically recover
				fmt.Printf("Kafka Error: %v\n", e)
			}
		}
	}

	fmt.Printf("Closing consumer topics: %s\n", topics)
	for _, c := range consumerOuput {
		close(c)
	}
	kc.Consumer.Close()
}
