package app

import (
	"context"
	"time"

	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
)

type App struct {
	logger  Logger
	storage storage.EventStorage
}

type Logger interface {
	Debug(msg string, a ...any)
	Info(msg string, a ...any)
	Warning(msg string, a ...any)
	Error(msg string, a ...any)
}

func New(logger Logger, storage storage.EventStorage) *App {
	return &App{
		logger:  logger,
		storage: storage,
	}
}

func (a *App) CreateEvent(ctx context.Context, event *storage.Event) error {
	return a.storage.CreateEvent(ctx, event)
}

func (a *App) UpdateEvent(ctx context.Context, eventID uuid.UUID, event *storage.Event) error {
	return a.storage.UpdateEvent(ctx, eventID, event)
}

func (a *App) DeleteEvent(ctx context.Context, eventID uuid.UUID) error {
	return a.storage.DeleteEvent(ctx, eventID)
}

func (a *App) GetEvents(ctx context.Context) ([]*storage.Event, error) {
	return a.storage.GetEvents(ctx)
}

func (a *App) GetEvent(ctx context.Context, eventID uuid.UUID) (*storage.Event, error) {
	return a.storage.GetEvent(ctx, eventID)
}

func (a *App) GetEventByDate(ctx context.Context, eventDatetime time.Time) (*storage.Event, error) {
	return a.storage.GetEventByDate(ctx, eventDatetime)
}

func (a *App) GetEventsForDay(ctx context.Context, startOfDay time.Time) ([]*storage.Event, error) {
	return a.storage.GetEventsForDay(ctx, startOfDay)
}

func (a *App) GetEventsForWeek(ctx context.Context, startOfWeek time.Time) ([]*storage.Event, error) {
	return a.storage.GetEventsForWeek(ctx, startOfWeek)
}

func (a *App) GetEventsForMonth(ctx context.Context, startOfMonth time.Time) ([]*storage.Event, error) {
	return a.storage.GetEventsForMonth(ctx, startOfMonth)
}
