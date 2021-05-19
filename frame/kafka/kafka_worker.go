package kafka

import (
	"errors"
)

var (
	gService *KafkaConsumer
)

func InitWorker(topic []string, groupId string, bootstrapServers string, securityProtocol string,
	saslMechanism string, saslUsername string, saslPassword string, handler KafkaHandler) error {
	gService = NewConsumer(topic, groupId, bootstrapServers, securityProtocol, saslMechanism,
		saslUsername, saslPassword, handler)

	if gService == nil {
		return errors.New("Unexpected error")
	}

	return nil
}
