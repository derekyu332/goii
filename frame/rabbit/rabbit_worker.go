package rabbit

import (
	"errors"
)

var (
	gService *RabbitConsumer
)

func InitWorker(amqpURI string, queueName string, key string, handler RabbitHandler) error {
	gService = NewConsumer(amqpURI, queueName, key, handler)

	if gService == nil {
		return errors.New("Unexpected error")
	}

	return nil
}
