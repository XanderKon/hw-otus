package rmq

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/streadway/amqp"
)

// Consumer ...
type Consumer struct {
	conn        *amqp.Connection
	channel     *amqp.Channel
	done        chan error
	consumerTag string

	uri          string
	exchangeName string
	exchangeType string
	queue        string
	bindingKey   string
	maxInterval  time.Duration

	MaxElapsedTime  time.Duration
	InitialInterval time.Duration
	Multiplier      float64
	MaxInterval     time.Duration
}

func NewConsumer(consumerTag, uri, exchangeName, exchangeType, queue, bindingKey string, maxInterval time.Duration) *Consumer {
	return &Consumer{
		consumerTag:  consumerTag,
		uri:          uri,
		exchangeName: exchangeName,
		exchangeType: exchangeType,
		queue:        queue,
		bindingKey:   bindingKey,
		done:         make(chan error),
		maxInterval:  maxInterval,
	}
}

type Worker func(context.Context, <-chan amqp.Delivery)

func (c *Consumer) Handle(ctx context.Context, fn Worker, threads int) error {
	var err error
	if err = c.connect(); err != nil {
		return fmt.Errorf("error: %v", err)
	}

	msgs, err := c.announceQueue()
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	for {
		for i := 0; i < threads; i++ {
			go fn(ctx, msgs)
		}

		if <-c.done != nil {
			msgs, err = c.reConnect(ctx)
			if err != nil {
				return fmt.Errorf("reconnecting Error: %s", err)
			}
		}
		fmt.Println("Reconnected... possibly")
	}
}

func (c *Consumer) Shutdown() error {
	// will close() the deliveries channel
	if err := c.channel.Cancel(c.consumerTag, true); err != nil {
		return fmt.Errorf("Consumer cancel failed: %s", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	return <-c.done
}

func (c *Consumer) connect() error {
	var err error

	c.conn, err = amqp.Dial(c.uri)
	if err != nil {
		return fmt.Errorf("dial: %s", err)
	}

	c.channel, err = c.conn.Channel()
	if err != nil {
		return fmt.Errorf("channel: %s", err)
	}

	go func() {
		log.Printf("closing: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
		// Понимаем, что канал сообщений закрыт, надо пересоздать соединение.
		c.done <- errors.New("channel Closed")
	}()

	if err = c.channel.ExchangeDeclare(
		c.exchangeName,
		c.exchangeType,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("exchange declare: %s", err)
	}

	return nil
}

// Задекларировать очередь, которую будем слушать.
func (c *Consumer) announceQueue() (<-chan amqp.Delivery, error) {
	queue, err := c.channel.QueueDeclare(
		c.queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("queue Declare: %s", err)
	}

	// Число сообщений, которые можно подтвердить за раз.
	err = c.channel.Qos(50, 0, false)
	if err != nil {
		return nil, fmt.Errorf("error setting qos: %s", err)
	}

	// Создаём биндинг (правило маршрутизации).
	if err = c.channel.QueueBind(
		queue.Name,
		c.bindingKey,
		c.exchangeName,
		false,
		nil,
	); err != nil {
		return nil, fmt.Errorf("queue Bind: %s", err)
	}

	msgs, err := c.channel.Consume(
		queue.Name,
		c.consumerTag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("queue Consume: %s", err)
	}

	return msgs, nil
}

func (c *Consumer) reConnect(ctx context.Context) (<-chan amqp.Delivery, error) {
	be := backoff.NewExponentialBackOff()
	be.MaxElapsedTime = c.MaxElapsedTime
	be.InitialInterval = c.InitialInterval
	be.Multiplier = c.Multiplier
	be.MaxInterval = c.maxInterval

	b := backoff.WithContext(be, ctx)
	for {
		d := b.NextBackOff()
		if d == backoff.Stop {
			return nil, fmt.Errorf("stop reconnecting")
		}

		select {
		case <-time.After(d):
			if err := c.connect(); err != nil {
				log.Printf("could not connect in reconnect call: %+v", err)
				continue
			}
			msgs, err := c.announceQueue()
			if err != nil {
				fmt.Printf("Couldn't connect: %+v", err)
				continue
			}

			return msgs, nil
		}
	}
}
