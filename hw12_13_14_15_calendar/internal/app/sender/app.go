package sender

import (
	"context"

	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/pkg/rmq"
	"github.com/streadway/amqp"
)

type Sender struct {
	logger  Logger
	rmq     *rmq.Rmq
	threads int
}

type Logger interface {
	Debug(msg string, a ...any)
	Info(msg string, a ...any)
	Warning(msg string, a ...any)
	Error(msg string, a ...any)
}

func New(
	logger Logger,
	rmq *rmq.Rmq,
	threads int,
) *Sender {
	return &Sender{
		logger:  logger,
		rmq:     rmq,
		threads: threads,
	}
}

func (s *Sender) Consume(ctx context.Context) error {
	return s.rmq.Handle(ctx, s.worker, s.threads)
}

func (s *Sender) worker(ctx context.Context, ch <-chan amqp.Delivery) {
	for {
		select {
		case msg := <-ch:
			s.logger.Info("successfully receive from queue: %s", msg.Body)
			msg.Ack(false)
		case <-ctx.Done():
			return
		}
	}
}
