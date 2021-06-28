package kafka

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	confluentKafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/lk153/go-lib/kafka/ccloud"
)

type KafkaProducer struct {
	ConfigFile *string
	conf       map[string]string
	producer   *confluentKafka.Producer
}

func (kp *KafkaProducer) InitConfig() error {
	if *kp.ConfigFile == "" {
		fmt.Printf("empty configFile\n")
		return fmt.Errorf("empty configFile")
	}

	kp.conf = ccloud.ReadCCloudConfig(*kp.ConfigFile)
	return nil
}

func (kp *KafkaProducer) CreateProducerInstance() error {
	switch {
	case kp.conf[ccloud.BOOTSTRAP_SERVERS] == "":
		return fmt.Errorf("miss BOOTSTRAP_SERVERS")
	}

	// Create Producer instance
	producer, err := confluentKafka.NewProducer(&confluentKafka.ConfigMap{
		// ccloud.METADATA_BROKER_LIST: kp.conf[ccloud.METADATA_BROKER_LIST],
		ccloud.BOOTSTRAP_SERVERS: kp.conf[ccloud.BOOTSTRAP_SERVERS],
		ccloud.SASL_MECHANISMS:   kp.conf[ccloud.SASL_MECHANISMS],
		ccloud.SECURITY_PROTOCOL: kp.conf[ccloud.SECURITY_PROTOCOL],
		ccloud.SASL_USERNAME:     kp.conf[ccloud.SASL_USERNAME],
		ccloud.SASL_PASSWORD:     kp.conf[ccloud.SASL_PASSWORD]})
	if err != nil {
		fmt.Printf("Failed to create producer: %s\n", err)
		return err
	}

	kp.producer = producer
	return nil
}

// CreateTopic creates a topic using the Admin Client API
func (kp *KafkaProducer) CreateTopic(topic string) {
	adminClient, err := confluentKafka.NewAdminClientFromProducer(kp.producer)
	if err != nil {
		fmt.Printf("Failed to create new admin client from producer: %s\n", err)
		os.Exit(1)
	}
	// Contexts are used to abort or limit the amount of time
	// the Admin call blocks waiting for a result.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Create topics on cluster.
	// Set Admin options to wait up to 60s for the operation to finish on the remote cluster
	maxDur, err := time.ParseDuration("60s")
	if err != nil {
		fmt.Printf("ParseDuration(60s): %s\n", err)
		os.Exit(1)
	}

	numPartitions, err := strconv.Atoi(kp.conf[ccloud.NUM_PARTITIONS])
	if err != nil {
		fmt.Printf("NUM_PARTITIONS ERROR: %s\n", err)
		numPartitions = 1
	}

	replicationFactor, err := strconv.Atoi(kp.conf[ccloud.REPLICATION_FACTOR])
	if err != nil {
		fmt.Printf("REPLICATION_FACTOR ERROR: %s\n", err)
		replicationFactor = 3
	}

	results, err := adminClient.CreateTopics(
		ctx,
		// Multiple topics can be created simultaneously
		// by providing more TopicSpecification structs here.
		[]confluentKafka.TopicSpecification{{
			Topic:             topic,
			NumPartitions:     numPartitions,
			ReplicationFactor: replicationFactor}},
		// Admin options
		confluentKafka.SetAdminOperationTimeout(maxDur),
	)

	if err != nil {
		fmt.Printf("Admin Client request error: %v\n", err)
		os.Exit(1)
	}

	for _, result := range results {
		if result.Error.Code() != confluentKafka.ErrNoError && result.Error.Code() != confluentKafka.ErrTopicAlreadyExists {
			fmt.Printf("Failed to create topic: %v\n", result.Error)
			os.Exit(1)
		}
		fmt.Printf("%v\n", result)
	}

	adminClient.Close()
}

func (kp *KafkaProducer) ProduceMessage(
	topic *string,
	recordValue string,
) error {
	doneChan := make(chan bool)
	go func() {
		defer close(doneChan)
		for e := range kp.producer.Events() {
			switch ev := e.(type) {
			case *confluentKafka.Message:
				m := ev
				if m.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
				} else {
					fmt.Printf("Delivered message to topic %s [%d] at offset %v\n",
						*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
				}
				return

			default:
				fmt.Printf("Ignored event: %s\n", ev)
			}
		}
	}()

	msg := &confluentKafka.Message{
		TopicPartition: confluentKafka.TopicPartition{
			Topic:     topic,
			Partition: confluentKafka.PartitionAny,
		}, Value: []byte(recordValue),
	}
	kp.producer.ProduceChannel() <- msg

	// wait for delivery report goroutine to finish
	<-doneChan

	return nil
}
