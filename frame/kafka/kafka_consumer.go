package kafka

import (
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/derekyu332/goii/helper/extend"
	"github.com/derekyu332/goii/helper/logger"
	"github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	logger           *zap.Logger
}

func NewConsumer(topic []string, groupId string, bootstrapServers string, securityProtocol string,
	saslMechanism string, saslUsername string, saslPassword string, handler KafkaHandler) *KafkaConsumer {
	logFile := "../output/app_%Y%m%d.log"
	rotator, err := rotatelogs.New(logFile, rotatelogs.WithMaxAge(60*24*time.Hour), rotatelogs.WithRotationTime(24*time.Hour))

	if err != nil {
		panic(err)
	}

	encoderConfig := map[string]string{
		"levelEncoder": "capital",
		"timeKey":      "date",
		"timeEncoder":  "iso8601",
	}
	data, _ := json.Marshal(encoderConfig)
	var encCfg zapcore.EncoderConfig

	if err := json.Unmarshal(data, &encCfg); err != nil {
		panic(err)
	}

	w := zapcore.AddSync(rotator)
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encCfg), w, zap.InfoLevel)
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
		logger:           zap.New(core),
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
			this.logger.Info("Got Message",
				zap.String("topic", *msg.TopicPartition.Topic),
				zap.Int32("partition", msg.TopicPartition.Partition),
				zap.Int64("offset", int64(msg.TopicPartition.Offset)),
				zap.String("content", extend.BytesString(msg.Value)))
			go this.handler(msg)
		} else {
			logger.Warning("ReadMessage error: %v (%v)\n", err, msg)
			consumer.Close()
			break
		}
	}

	defer this.logger.Sync()
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
