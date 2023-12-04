package storage

import (
	"context"

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
	// GetEventsByDay(ctx context.Context, startOfDay time.Time) ([]*Event, error)
	// GetEventsByWeek(ctx context.Context, startOfWeek time.Time) ([]*Event, error)
	// GetEventsByMonth(ctx context.Context, startOfMonth time.Time) ([]*Event, error)
}
