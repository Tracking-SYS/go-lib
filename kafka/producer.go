package kafka

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/lk153/go-lib/kafka/ccloud"
)

type KafkaProducer struct {
	ConfigFile *string
	conf       map[string]string
}

func (kp *KafkaProducer) InitConfig() error {
	if *kp.ConfigFile == "" {
		fmt.Printf("empty configFile")
		return fmt.Errorf("empty configFile")
	}

	kp.conf = ccloud.ReadCCloudConfig(*kp.ConfigFile)
	return nil
}

// CreateTopic creates a topic using the Admin Client API
func CreateTopic(p *kafka.Producer, topic string) {
	adminClient, err := kafka.NewAdminClientFromProducer(p)
	if err != nil {
		fmt.Printf("Failed to create new admin client from producer: %s", err)
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
		fmt.Printf("ParseDuration(60s): %s", err)
		os.Exit(1)
	}
	results, err := adminClient.CreateTopics(
		ctx,
		// Multiple topics can be created simultaneously
		// by providing more TopicSpecification structs here.
		[]kafka.TopicSpecification{{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 3}},
		// Admin options
		kafka.SetAdminOperationTimeout(maxDur))
	if err != nil {
		fmt.Printf("Admin Client request error: %v\n", err)
		os.Exit(1)
	}
	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError && result.Error.Code() != kafka.ErrTopicAlreadyExists {
			fmt.Printf("Failed to create topic: %v\n", result.Error)
			os.Exit(1)
		}
		fmt.Printf("%v\n", result)
	}
	adminClient.Close()

}

func (kp *KafkaProducer) ProduceMessage(
	configFile *string,
	topic *string,
	recordKey string,
	recordValue string,
) error {
	producer, err := kp.createProducerInstance()
	if err != nil {
		return err
	}

	// Create topic if needed
	CreateTopic(producer, *topic)

	deliveryChan := make(chan kafka.Event)
	err = pushMessage(producer, topic, recordKey, recordValue, deliveryChan)
	if err != nil {
		return err
	}

	err = checkMessageDeliver(deliveryChan)
	if err != nil {
		return err
	}

	return nil
}

func (kp *KafkaProducer) createProducerInstance() (*kafka.Producer, error) {
	// Create Producer instance
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": kp.conf["bootstrap.servers"],
		"sasl.mechanisms":   kp.conf["sasl.mechanisms"],
		"security.protocol": kp.conf["security.protocol"],
		"sasl.username":     kp.conf["sasl.username"],
		"sasl.password":     kp.conf["sasl.password"]})
	if err != nil {
		fmt.Printf("Failed to create producer: %s", err)
		return nil, err
	}

	return producer, nil
}

func pushMessage(
	producer *kafka.Producer,
	topic *string,
	recordKey string,
	recordValue string,
	deliveryChan chan kafka.Event,
) error {
	err := producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: topic, Partition: int32(kafka.PartitionAny)},
		Key:            []byte(recordKey),
		Value:          []byte(recordValue),
	}, deliveryChan)

	if err != nil {
		fmt.Printf("Produce message has error: key %v, val %v", recordKey, recordValue)
		return err
	}

	return nil
}

func checkMessageDeliver(deliveryChan chan kafka.Event) error {
	kafkaEvent := <-deliveryChan
	msg := kafkaEvent.(*kafka.Message)

	if msg.TopicPartition.Error != nil {
		fmt.Printf("Delivery failed: %v\n", msg.TopicPartition.Error)
		return msg.TopicPartition.Error
	} else {
		fmt.Printf("Delivered message to topic %s [%d] at offset %v\n",
			*msg.TopicPartition.Topic, msg.TopicPartition.Partition, msg.TopicPartition.Offset)
	}

	close(deliveryChan)
	return nil
}
