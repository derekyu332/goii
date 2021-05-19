package kafka

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/derekyu332/goii/helper/logger"
	"time"
)

const (
	KAFKA_RECONNECT_DELAY = 5 * time.Second
)

type KafkaHandler func(message *kafka.Message) error

type KafkaConsumer struct {
	Topic            []string
	GroupId          string
	BootstrapServers string
	SecurityProtocol string
	SaslMechanism    string
	SaslUsername     string
	SaslPassword     string
	done             chan bool
	handler          KafkaHandler
}

func NewConsumer(topic []string, groupId string, bootstrapServers string, securityProtocol string,
	saslMechanism string, saslUsername string, saslPassword string, handler KafkaHandler) *KafkaConsumer {
	var err error
	newConsumer := &KafkaConsumer{
		Topic:            topic,
		GroupId:          groupId,
		BootstrapServers: bootstrapServers,
		SecurityProtocol: securityProtocol,
		SaslMechanism:    saslMechanism,
		SaslUsername:     saslUsername,
		SaslPassword:     saslPassword,
		done:             make(chan bool),
		handler:          handler,
	}

	err = newConsumer.initConsumer()

	if err != nil {
		panic(err)
	}

	return newConsumer
}

func (this *KafkaConsumer) initConsumer() error {
	logger.Info("Init Kafka Consumer, it may take a few seconds to init the connection\n")

	var kafkaconf = &kafka.ConfigMap{
		"api.version.request":       "true",
		"auto.offset.reset":         "latest",
		"heartbeat.interval.ms":     3000,
		"session.timeout.ms":        30000,
		"max.poll.interval.ms":      120000,
		"fetch.max.bytes":           1024000,
		"max.partition.fetch.bytes": 256000}
	kafkaconf.SetKey("bootstrap.servers", this.BootstrapServers)
	kafkaconf.SetKey("group.id", this.GroupId)

	switch this.SecurityProtocol {
	case "plaintext":
		kafkaconf.SetKey("security.protocol", "plaintext")
	case "sasl_ssl":
		kafkaconf.SetKey("security.protocol", "sasl_ssl")
		kafkaconf.SetKey("ssl.ca.location", "./conf/ca-cert.pem")
		kafkaconf.SetKey("sasl.username", this.SaslUsername)
		kafkaconf.SetKey("sasl.password", this.SaslPassword)
		kafkaconf.SetKey("sasl.mechanism", this.SaslMechanism)
	case "sasl_plaintext":
		kafkaconf.SetKey("security.protocol", "sasl_plaintext")
		kafkaconf.SetKey("sasl.username", this.SaslUsername)
		kafkaconf.SetKey("sasl.password", this.SaslPassword)
		kafkaconf.SetKey("sasl.mechanism", this.SaslMechanism)

	default:
		return kafka.NewError(kafka.ErrUnknownProtocol, "unknown protocol", true)
	}

	consumer, err := kafka.NewConsumer(kafkaconf)

	if err != nil {
		return err
	}

	logger.Warning("Init Kafka Consumer success\n")
	consumer.SubscribeTopics(this.Topic, nil)
	go this.consumeMessage(consumer)

	return nil
}

func (this *KafkaConsumer) consumeMessage(consumer *kafka.Consumer) bool {
	for {
		msg, err := consumer.ReadMessage(-1)

		if err == nil {
			logger.Info("Got Message: [%v] %v", msg.TopicPartition, string(msg.Value))
		} else {
			logger.Warning("ReadMessage error: %v (%v)\n", err, msg)
			consumer.Close()
			break
		}
	}

	logger.Warning("Consumer Reconnect")

	for {
		select {
		case <-this.done:
			consumer.Close()
			logger.Warning("Consumer Closed.")
			return false
		case <-time.After(KAFKA_RECONNECT_DELAY):
			logger.Warning("Consumer Retrying Init...")
			if this.initConsumer() == nil {
				return true
			}
		}
	}

	return true
}
