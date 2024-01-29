package rmq

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/streadway/amqp"
)

var (
	ErrChannelClosed = errors.New("channel Closed")
	ErrChannel       = errors.New("channel error")
	ErrWithQueue     = errors.New("errors with Queue")
	ErrGeneralError  = errors.New("there is a some problem with RMQ server")
	ErrReconnection  = errors.New("there is a some problem with reconnect to RMQ server")
	ErrClose         = errors.New("AMQP connection close error")
	ErrConnections   = errors.New("can't connect to the RMQ server")
	ErrPublish       = errors.New("AMQP publish error")
)

// Consumer ...
type Rmq struct {
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

func NewRmq(
	consumerTag string,
	uri string,
	exchangeName string,
	exchangeType string,
	queue string,
	bindingKey string,
	maxInterval time.Duration,
) *Rmq {
	return &Rmq{
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

func (r *Rmq) Connect() error {
	var err error
	if err = r.connect(); err != nil {
		return errors.Join(ErrGeneralError, err)
	}

	err = r.announceQueue()
	if err != nil {
		return errors.Join(ErrWithQueue, err)
	}

	return nil
}

func (r *Rmq) Handle(ctx context.Context, fn Worker, threads int) error {
	msgs, err := r.consume()
	if err != nil {
		return errors.Join(ErrGeneralError, err)
	}

	for {
		for i := 0; i < threads; i++ {
			go fn(ctx, msgs)
		}

		select {
		case <-ctx.Done():
			return nil
		case <-r.done:
			err = r.reConnect(ctx)
			if err != nil {
				return errors.Join(ErrReconnection, err)
			}
		}
	}
}

func (r *Rmq) Shutdown() error {
	// will close() the deliveries channel
	if err := r.channel.Cancel(r.consumerTag, true); err != nil {
		return errors.Join(ErrGeneralError, err)
	}

	if err := r.conn.Close(); err != nil {
		return errors.Join(ErrClose, err)
	}

	return <-r.done
}

func (r *Rmq) Publish(msg amqp.Publishing) error {
	if r.channel == nil {
		return nil
	}
	if err := r.channel.Publish(r.exchangeName, r.queue, false, false, msg); err != nil {
		return errors.Join(ErrPublish, err)
	}

	return nil
}

func (r *Rmq) connect() error {
	var err error

	r.conn, err = amqp.Dial(r.uri)
	if err != nil {
		return errors.Join(ErrConnections, err)
	}

	r.channel, err = r.conn.Channel()
	if err != nil {
		return errors.Join(ErrChannel, err)
	}

	go func() {
		<-r.conn.NotifyClose(make(chan *amqp.Error))
		// Понимаем, что канал сообщений закрыт, надо пересоздать соединение.
		r.done <- ErrChannelClosed
	}()

	if err = r.channel.ExchangeDeclare(
		r.exchangeName,
		r.exchangeType,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return errors.Join(ErrGeneralError, err)
	}

	return nil
}

// Задекларировать очередь, которую будем слушать.
func (r *Rmq) announceQueue() error {
	queue, err := r.channel.QueueDeclare(
		r.queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return errors.Join(ErrGeneralError, err)
	}

	// Число сообщений, которые можно подтвердить за раз.
	err = r.channel.Qos(50, 0, false)
	if err != nil {
		return errors.Join(ErrGeneralError, err)
	}

	// Создаём биндинг (правило маршрутизации).
	if err = r.channel.QueueBind(
		queue.Name,
		r.bindingKey,
		r.exchangeName,
		false,
		nil,
	); err != nil {
		return errors.Join(ErrGeneralError, err)
	}

	return nil
}

func (r *Rmq) consume() (<-chan amqp.Delivery, error) {
	msgs, err := r.channel.Consume(
		r.queue,
		r.consumerTag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, errors.Join(ErrGeneralError, err)
	}

	return msgs, nil
}

func (r *Rmq) reConnect(ctx context.Context) error {
	be := backoff.NewExponentialBackOff()
	be.MaxElapsedTime = r.MaxElapsedTime
	be.InitialInterval = r.InitialInterval
	be.Multiplier = r.Multiplier
	be.MaxInterval = r.maxInterval

	b := backoff.WithContext(be, ctx)
	for {
		d := b.NextBackOff()
		if d == backoff.Stop {
			return ErrReconnection
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(d):
			if err := r.connect(); err != nil {
				fmt.Printf("could not connect in reconnect call: %+v", err)
				continue
			}
			err := r.announceQueue()
			if err != nil {
				fmt.Printf("Couldn't connect: %+v", err)
				continue
			}

			return nil
		}
	}
}
