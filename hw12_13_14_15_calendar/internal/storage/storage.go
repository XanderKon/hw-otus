package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type EventStorage interface {
	Connect(ctx context.Context) error
	Close() error
	CreateEvent(ctx context.Context, event *Event) error
	UpdateEvent(ctx context.Context, eventID uuid.UUID, event *Event) error
	DeleteEvent(ctx context.Context, eventID uuid.UUID) error
	GetEvents(ctx context.Context) ([]*Event, error)
	GetEvent(ctx context.Context, eventID uuid.UUID) (*Event, error)
	GetEventByDate(ctx context.Context, eventDatetime time.Time) (*Event, error)
	GetEventsForDay(ctx context.Context, startOfDay time.Time) ([]*Event, error)
	GetEventsForWeek(ctx context.Context, startOfWeek time.Time) ([]*Event, error)
	GetEventsForMonth(ctx context.Context, startOfMonth time.Time) ([]*Event, error)
	GetEventsForNotifications(ctx context.Context) ([]*Event, error)
	DeleteOldEvents(ctx context.Context, duration time.Duration) (int, error)
}
