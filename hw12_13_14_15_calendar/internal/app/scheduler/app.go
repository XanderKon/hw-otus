package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/storage"
	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/pkg/rmq"
	"github.com/streadway/amqp"
)

var (
	ErrGetEventsForNotification = errors.New("cannot get events for notification")
	ErrSerializeNotification    = errors.New("can't serizlize notification object")
	ErrSendNotificationToQueue  = errors.New("can't send notification to queue")
)

type Scheduler struct {
	logger                 Logger
	storage                storage.EventStorage
	rmq                    *rmq.Rmq
	runFrequencyInterval   time.Duration
	timeForRemoveOldEvents time.Duration
}

type Logger interface {
	Debug(msg string, a ...any)
	Info(msg string, a ...any)
	Warning(msg string, a ...any)
	Error(msg string, a ...any)
}

func New(
	logger Logger,
	storage storage.EventStorage,
	rmq *rmq.Rmq,
	runFrequencyInterval time.Duration,
	timeForRemoveOldEvents time.Duration,
) *Scheduler {
	return &Scheduler{
		logger:                 logger,
		storage:                storage,
		rmq:                    rmq,
		runFrequencyInterval:   runFrequencyInterval,
		timeForRemoveOldEvents: timeForRemoveOldEvents,
	}
}

func (s *Scheduler) NotificationSender(ctx context.Context) {
	ticker := time.NewTicker(s.runFrequencyInterval)
	s.logger.Info("successfully init timer")
	go func() {
		for {
			select {
			case <-ticker.C:
				// put to queue
				err := s.putNotificationsToQueue(ctx)
				if err != nil {
					s.logger.Error("put to queue error: %w", err)
				}

				// delete old events
				err = s.deleteOldEvents(ctx)
				if err != nil {
					s.logger.Error("delete old events error: %w", err)
				}
			case <-ctx.Done():
				ticker.Stop()
				s.logger.Info("successfully stop timer")
				return
			}
		}
	}()
}

func (s *Scheduler) putNotificationsToQueue(ctx context.Context) error {
	events, err := s.getEventsForNotifications(ctx)
	if err != nil {
		return errors.Join(err, ErrGetEventsForNotification)
	}

	for _, event := range events {
		notification := s.getNotificationForEvent(event)

		data, err := json.Marshal(notification)
		if err != nil {
			return errors.Join(err, ErrSerializeNotification)
		}

		err = s.sendToQueue(data)
		if err != nil {
			return errors.Join(err, ErrSendNotificationToQueue)
		}

		event.NotifyAt = time.Now()
		err = s.setNotifyTimeForEvent(ctx, event)
		if err != nil {
			return errors.Join(err, ErrSendNotificationToQueue)
		}

		s.logger.Debug("successfully put notification to queue: %s", data)
	}

	return nil
}

func (s *Scheduler) sendToQueue(data []byte) error {
	msg := amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	}
	return s.rmq.Publish(msg)
}

func (s *Scheduler) getEventsForNotifications(ctx context.Context) ([]*storage.Event, error) {
	return s.storage.GetEventsForNotifications(ctx)
}

func (s *Scheduler) setNotifyTimeForEvent(ctx context.Context, event *storage.Event) error {
	return s.storage.UpdateEvent(ctx, event.ID, event)
}

func (s *Scheduler) deleteOldEvents(ctx context.Context) error {
	count, err := s.storage.DeleteOldEvents(ctx, s.timeForRemoveOldEvents)
	if err != nil {
		return err
	}

	if count > 0 {
		s.logger.Debug("successfully remove %d old events", count)
	}
	return nil
}

func (s *Scheduler) getNotificationForEvent(event *storage.Event) *storage.Notification {
	return &storage.Notification{
		EventID:  event.ID.String(),
		Title:    event.Title,
		DateTime: event.DateTime,
		UserID:   event.UserID,
	}
}
