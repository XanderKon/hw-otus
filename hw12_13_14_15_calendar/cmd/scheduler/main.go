package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/logger"
	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/rmq"
	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/storage/sql"
	"github.com/streadway/amqp"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/scheduler_config.toml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	// init context
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	config := NewConfig()

	logg := logger.New(config.Logger.Level, os.Stdout)

	var eventStorage storage.EventStorage
	if config.Storage.Driver == "postgres" {
		connectionString := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			config.DB.DBHost, config.DB.DBPort, config.DB.DBUsername, config.DB.DBPassword, config.DB.DBName,
		)

		eventStorage = sqlstorage.New(connectionString)
		err := eventStorage.Connect(ctx)
		if err != nil {
			logg.Error("cannot connect to DB server: " + err.Error())
			cancel()
			os.Exit(1) //nolint:gocritic
		}
		defer eventStorage.Close()
	} else {
		eventStorage = memorystorage.New()
	}

	logg.Info(fmt.Sprintf("successfully init %s storage", config.Storage.Driver))

	// conn, err := amqp.Dial(config.rmq.Uri)

	// if err != nil {
	// 	logg.Error("cannot connect to RMQ server: " + err.Error())
	// 	cancel()
	// 	os.Exit(1) //nolint:gocritic
	// }

	c := rmq.NewConsumer(
		config.Rmq.ConsumerTag,
		config.Rmq.URI,
		config.Rmq.Exchange.Name,
		config.Rmq.Exchange.Type,
		config.Rmq.Exchange.QueueName,
		config.Rmq.Exchange.BindingKey,
		config.Rmq.MaxInterval,
	)

	go func() {
		<-ctx.Done()

		_, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := c.Shutdown(); err != nil {
			logg.Error("failed to shutdown RMQ server: " + err.Error())
		}

		logg.Info("RMQ server successfully terminated!")
	}()

	err := c.Handle(ctx, Worker, 1)
	if err != nil {
		fmt.Println(err)
	}
}

func Worker(ctx context.Context, ch <-chan amqp.Delivery) {
	fmt.Println(<-ch)
}
