package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/app/scheduler"
	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/logger"
	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/storage/sql"
	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/pkg/rmq"
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
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGTSTP)
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
		cancel()
		os.Exit(1)
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

	scheduler := scheduler.New(logg, eventStorage, rmqInstance, config.Scheduler.RunFrequencyInterval, config.Scheduler.TimeForRemoveOldEvents)

	wg.Add(1)
	go func() {
		scheduler.NotificationSender(ctx)
	}()
	wg.Wait()
}

// func Worker(ctx context.Context, ch <-chan amqp.Delivery) {
// 	for msg := range ch {
// 		fmt.Printf("Received a message: %s\n", string(msg.Body))
// 		msg.Ack(false)
// 	}
// }
