package rabbit

import (
	"errors"
	"github.com/derekyu332/goii/helper/logger"
	"github.com/streadway/amqp"
	"sync"
	"time"
)

type message []byte
type RabbitHandler func(amqp.Delivery) error

type RabbitConsumer struct {
	uri           string
	queueName     string
	routineKey    string
	session       *RabbitSession
	channel       *amqp.Channel
	done          chan bool
	callbacks     map[string]chan<- message
	gCallbackLock sync.RWMutex
	handler       RabbitHandler
}

func NewConsumer(amqpURI string, queueName string, key string, handler RabbitHandler) *RabbitConsumer {
	var err error
	newConsumer := &RabbitConsumer{
		uri:        amqpURI,
		queueName:  queueName,
		routineKey: key,
		done:       make(chan bool),
		callbacks:  make(map[string]chan<- message),
		handler:    handler,
	}
	newConsumer.session = NewSession(amqpURI)

	if err != nil {
		panic(err)
	}

	err = newConsumer.initConsumer()

	if err != nil {
		panic(err)
	}

	return newConsumer
}

func (this *RabbitConsumer) Connection() *amqp.Connection {
	return this.session.Connection()
}

func (this *RabbitConsumer) AddCallback(correlationId string, callback chan<- message) {
	this.gCallbackLock.Lock()
	this.callbacks[correlationId] = callback
	this.gCallbackLock.Unlock()
}

func (this *RabbitConsumer) DelCallback(correlationId string) {
	this.gCallbackLock.Lock()
	delete(this.callbacks, correlationId)
	this.gCallbackLock.Unlock()
}

func (this *RabbitConsumer) initConsumer() error {
	if this.session.Connection() == nil || this.session.Connection().IsClosed() {
		return errors.New("Connection Lost")
	}

	var err error
	this.channel, err = this.session.Connection().Channel()

	if err != nil {
		return err
	}

	logger.Warning("Open Consumer Channel Success")

	if err = this.channel.QueueBind(
		this.queueName, this.routineKey,
		"amq.direct",
		false,
		nil); err != nil {
		return err
	}

	logger.Warning("Consumer Queue %v Bind %v Success", this.queueName, this.routineKey)
	go this.subscribe()
	return nil
}

func (this *RabbitConsumer) subscribe() error {
	deliveries, err := this.channel.Consume(
		this.queueName,            // name
		"consumer"+this.queueName, // consumerTag,
		true,  // noAck
		true,  // exclusive
		false, // noLocal
		false, // noWait
		nil,   // arguments
	)

	if err != nil {
		logger.Warning("Consume failed %v", err.Error())

		return err
	}

	logger.Warning("Consumer Queue %v Consume Success", this.queueName)
	go this.consumeMessage(deliveries)

	return nil
}

func (this *RabbitConsumer) consumeMessage(deliveries <-chan amqp.Delivery) bool {
	for d := range deliveries {
		logger.Info("Got %dB Delivery: [%v] %v",
			len(d.Body), d.DeliveryTag, d.CorrelationId)

		if len(d.Body) > 0 && d.CorrelationId != "" {
			cor := string([]byte(d.CorrelationId)[:10])
			this.gCallbackLock.Lock()

			if ch, ok := this.callbacks[cor]; ok {
				ch <- d.Body
				delete(this.callbacks, cor)
			}

			this.gCallbackLock.Unlock()
		} else if this.handler != nil {
			go this.handler(d)
		}
	}

	logger.Warning("Consumer Lost")

	for {
		select {
		case <-this.done:
			logger.Warning("Consumer Closed.")
			return false
		case <-time.After(RABBIT_RECONNECT_DELAY):
			logger.Warning("Consumer Retrying Init...")
			if this.initConsumer() == nil {
				return true
			}
		}
	}

	return true
}
