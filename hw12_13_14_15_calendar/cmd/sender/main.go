package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/app/sender"
	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/logger"
	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/pkg/rmq"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/sender_config.toml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	// init context
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGTSTP)
	defer cancel()

	config := NewConfig()

	logg := logger.New(config.Logger.Level, os.Stdout)

	rmqInstance := rmq.NewRmq(
		config.Rmq.ConsumerTag,
		config.Rmq.URI,
		config.Rmq.Exchange.Name,
		config.Rmq.Exchange.Type,
		config.Rmq.Exchange.QueueName,
		config.Rmq.Exchange.BindingKey,
		config.Rmq.MaxInterval,
	)

	err := rmqInstance.Connect()
	if err != nil {
		logg.Error("cannot connect to AMQP server: " + err.Error())
		return
	}

	var wg sync.WaitGroup

	go func() {
		<-ctx.Done()

		_, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := rmqInstance.Shutdown(); err != nil && !errors.Is(err, rmq.ErrChannelClosed) {
			logg.Error("failed to shutdown RMQ server: " + err.Error())
		}

		logg.Info("RMQ server successfully terminated!")

		wg.Done()
	}()

	sender := sender.New(logg, rmqInstance, config.Sender.Threads)

	wg.Add(1)
	go func() {
		err := sender.Consume(ctx)
		if err != nil {
			wg.Done()
			logg.Error("cannot init conumer for AMQP server: %s", err.Error())
		}
	}()
	wg.Wait()
}
