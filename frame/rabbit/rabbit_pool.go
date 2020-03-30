package rabbit

import (
	"errors"
	"github.com/derekyu332/goii/helper/logger"
	"github.com/silenceper/pool"
	"github.com/streadway/amqp"
)

var (
	gRabbitSession *RabbitSession
)

const (
	RABBIT_INIT_POOL_SIZE = 8
	RABBIT_MAX_POOL_SIZE  = 32
)

func NewPool(amqpURI string) pool.Pool {
	gRabbitSession = NewSession(amqpURI)

	if gRabbitSession == nil {
		return nil
	}

	factory := func() (interface{}, error) {
		if gRabbitSession.Connection() == nil || gRabbitSession.Connection().IsClosed() {
			return nil, errors.New("Connection Lost")
		}

		logger.Info("Create Rabbit Channel")
		ch := gRabbitSession.Channel()

		if ch == nil {
			return nil, errors.New("Connection Lost")
		} else {
			return ch, nil
		}
	}

	close := func(v interface{}) error {
		logger.Info("Rabbit Channel Closed")
		return v.(*amqp.Channel).Close()
	}

	ping := func(v interface{}) error {
		if gRabbitSession.IsChannelValid(v.(*amqp.Channel)) {
			return nil
		} else {
			return errors.New("Channel Invalid")
		}
	}

	poolConfig := &pool.Config{
		InitialCap:  RABBIT_INIT_POOL_SIZE,
		MaxIdle:     RABBIT_INIT_POOL_SIZE,
		MaxCap:      RABBIT_MAX_POOL_SIZE,
		Factory:     factory,
		Close:       close,
		Ping:        ping,
		IdleTimeout: 0,
	}

	pool, err := pool.NewChannelPool(poolConfig)

	if err != nil {
		panic(err)
	}

	return pool
}
