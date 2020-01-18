package rabbit

import (
	"github.com/derekyu332/goii/helper/logger"
	"github.com/streadway/amqp"
	"time"
)

const (
	RABBIT_RECONNECT_DELAY = 5 * time.Second
)

type RabbitSession struct {
	uri        string
	connection *amqp.Connection
	done       chan bool
	channels   []*amqp.Channel
}

func NewSession(amqpURI string) *RabbitSession {
	logger.Warning("Try Connect...")
	var err error
	newConnection := &RabbitSession{
		uri:      amqpURI,
		done:     make(chan bool),
		channels: make([]*amqp.Channel, 0),
	}
	newConnection.connection, err = amqp.Dial(newConnection.uri)

	if err != nil {
		panic(err)
	}

	logger.Warning("Connect Success")
	go newConnection.handelCheck()

	return newConnection
}

func (this *RabbitSession) Connection() *amqp.Connection {
	return this.connection
}

func (this *RabbitSession) Channel() *amqp.Channel {
	if this.connection == nil || this.connection.IsClosed() {
		return nil
	}

	if ch, err := this.connection.Channel(); ch != nil && err == nil {
		this.channels = append(this.channels, ch)

		return ch
	} else {
		return nil
	}
}

func (this *RabbitSession) IsChannelValid(ch *amqp.Channel) bool {
	if this.connection == nil || this.connection.IsClosed() {
		return false
	}

	for _, c := range this.channels {
		if ch == c {
			return true
		}
	}

	return false
}

func (this *RabbitSession) Close() {
	close(this.done)
}

func (this *RabbitSession) handelCheck() {
	for {
		select {
		case <-this.connection.NotifyClose(make(chan *amqp.Error)):
			logger.Warning("Connection Lost. ReConnect...")
			this.handelReconnect()
		case <-this.done:
			logger.Warning("Connection Closed")
			return
		}
	}
}

func (this *RabbitSession) handelReconnect() bool {
	for {
		var err error
		this.channels = make([]*amqp.Channel, 0)
		this.connection, err = amqp.Dial(this.uri)

		if err != nil {
			logger.Warning("Failed to Connect. Retrying...")

			select {
			case <-this.done:
				return false
			case <-time.After(RABBIT_RECONNECT_DELAY):
			}
		} else {
			logger.Warning("Connection ReConnect Success.")
			break
		}
	}

	return true
}
