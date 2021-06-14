package kafka

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/lk153/go-lib/kafka/ccloud"
)

// ConsumeRecordValue represents the struct of the value in a Kafka message
type ConsumeRecordValue interface{}

func Start(consumerOuput chan []byte) {
	// Initialization
	configFile, topic := ccloud.ParseArgs()
	conf := ccloud.ReadCCloudConfig(*configFile)

	fmt.Print(conf[ccloud.BOOTSTRAP_SERVERS])
	// Create Consumer instance
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		ccloud.METADATA_BROKER_LIST:       conf[ccloud.METADATA_BROKER_LIST],
		ccloud.BOOTSTRAP_SERVERS:          conf[ccloud.BOOTSTRAP_SERVERS],
		ccloud.SASL_MECHANISMS:            conf[ccloud.SASL_MECHANISMS],
		ccloud.SECURITY_PROTOCOL:          conf[ccloud.SECURITY_PROTOCOL],
		ccloud.SASL_USERNAME:              conf[ccloud.SASL_USERNAME],
		ccloud.SASL_PASSWORD:              conf[ccloud.SASL_PASSWORD],
		ccloud.GROUP_ID:                   conf[ccloud.GROUP_ID],
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		"enable.partition.eof":            true,
		"auto.offset.reset":               "earliest",
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
		case ev := <-c.Events():
			switch e := ev.(type) {
			case kafka.AssignedPartitions:
				fmt.Fprintf(os.Stderr, "%% %v\n", e)
				err = c.Assign(e.Partitions)
				if err != nil {
					fmt.Printf("Assign has error: %v", err)
				}
			case kafka.RevokedPartitions:
				fmt.Fprintf(os.Stderr, "%% %v\n", e)
				err = c.Unassign()
				if err != nil {
					fmt.Printf("Unassign has error: %v", err)
				}
			case *kafka.Message:
				fmt.Printf("%% Message on %s:\n%s\n",
					e.TopicPartition, string(e.Value))
				consumerOuput <- e.Value
			case kafka.PartitionEOF:
				fmt.Printf("%% Reached %v\n", e)
			case kafka.Error:
				// Errors should generally be considered as informational, the client will try to automatically recover
				fmt.Fprintf(os.Stderr, "%% Error: %v\n", e)
			}
		}
	}

	fmt.Printf("Closing consumer\n")
	c.Close()
}
