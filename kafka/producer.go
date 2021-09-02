package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Tracking-SYS/go-lib/kafka/ccloud"
	"github.com/Tracking-SYS/go-lib/slice"
	confluentKafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

var excludeConfig = []string{
	"num.partitions",
	"replication.factor",
}

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
	kafkaConfig := &confluentKafka.ConfigMap{
		"session.timeout.ms":   10000,
		"default.topic.config": confluentKafka.ConfigMap{"auto.offset.reset": "earliest"},
	}

	for k, v := range kp.conf {
		if v == "" || slice.Contains(excludeConfig, k) {
			continue
		}

		err := kafkaConfig.SetKey(k, v)
		if err != nil {
			fmt.Printf("Failed to set key: %s\n", err)
			os.Exit(1)
		}
	}

	kafkaConfigJSON, err := json.Marshal(kafkaConfig)
	if err != nil {
		fmt.Printf("Error: %s", err)
		return err
	}
	fmt.Printf("%v\n", string(kafkaConfigJSON))
	// Create Producer instance
	producer, err := confluentKafka.NewProducer(kafkaConfig)
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

	numPartitions := 1
	if kp.conf[ccloud.NUM_PARTITIONS] != "" {
		numPartitions, err = strconv.Atoi(kp.conf[ccloud.NUM_PARTITIONS])
		if err != nil {
			fmt.Printf("NUM_PARTITIONS ERROR: %s\n", err)
		}
	}

	replicationFactor := 3
	if kp.conf[ccloud.REPLICATION_FACTOR] != "" {
		replicationFactor, err = strconv.Atoi(kp.conf[ccloud.REPLICATION_FACTOR])
		if err != nil {
			fmt.Printf("REPLICATION_FACTOR ERROR: %s\n", err)
		}
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
