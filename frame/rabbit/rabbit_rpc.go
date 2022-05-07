package rabbit

import (
	"errors"
	"github.com/derekyu332/goii/helper/logger"
	"github.com/silenceper/pool"
	"github.com/streadway/amqp"
	"strconv"
	"sync/atomic"
	"time"
)

var (
	gProducerPool pool.Pool
	gConsumer     *RabbitConsumer
	gCorrId       int64
)

func InitProducer(amqpURI string, initCap int, maxCap int, maxIdle int) error {
	gProducerPool = NewPool(amqpURI, initCap, maxCap, maxIdle)

	if gProducerPool == nil {
		return errors.New("Unexpected error")
	}

	logger.Warning("Rabbit Producer Init Success")

	return nil
}

func InitConsumer(amqpURI string, queueName string, key string) error {
	gConsumer = NewConsumer(amqpURI, queueName, key, nil)

	if gConsumer == nil {
		return errors.New("Unexpected error")
	}

	gCorrId = time.Now().Unix()
	logger.Warning("Rabbit Consumer Init Success")

	return nil
}

func GetProducerChannel() *amqp.Channel {
	ch, err := gProducerPool.Get()

	if err != nil {
		logger.Warning("Get Channel Failed %v", err.Error())
		return nil
	}

	channel, ok := ch.(*amqp.Channel)

	if ok {
		return channel
	} else {
		return nil
	}
}

func RPC(body []byte, routineKey string, timeout time.Duration) ([]byte, error) {
	if gConsumer == nil || gConsumer.Connection() == nil || gConsumer.Connection().IsClosed() {
		return nil, errors.New("Unexpected error")
	}

	callback := make(chan message)
	curCorrId := atomic.AddInt64(&gCorrId, 1)
	CorrelationId := strconv.FormatInt(curCorrId, 10)
	gConsumer.AddCallback(CorrelationId, callback)
	defer gConsumer.DelCallback(CorrelationId)
	ch := GetProducerChannel()

	if ch == nil {
		return nil, errors.New("Unexpected error")
	}

	defer gProducerPool.Put(ch)

	if err := ch.Publish(
		"amq.direct",
		routineKey,
		false,
		false,
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			CorrelationId:   CorrelationId,
			Body:            body,
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
		},
	); err != nil {
		return nil, err
	}

	for {
		select {
		case response := <-callback:
			return []byte(response), nil
		case <-time.After(timeout):
			return nil, errors.New("Timeout")
		}
	}

	return nil, errors.New("Unexpected error")
}

func Notify(message []byte, routineKey string) error {
	ch := GetProducerChannel()

	if ch == nil {
		return errors.New("Unexpected error")
	}

	defer gProducerPool.Put(ch)

	if err := ch.Publish(
		"amq.direct",
		routineKey,
		false,
		false,
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            message,
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
		},
	); err != nil {
		return err
	}

	return nil
}
