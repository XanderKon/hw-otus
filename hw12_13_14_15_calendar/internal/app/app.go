package app

import (
	"context"

	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/storage"
)

type App struct {
	logger  Logger
	storage storage.EventStorage
}

type Logger interface { // TODO
}

func New(logger Logger, storage storage.EventStorage) *App {
	return &App{
		logger:  logger,
		storage: storage,
	}
}

func (a *App) CreateEvent(ctx context.Context, event storage.Event) error {
	return a.storage.CreateEvent(ctx, &event)
}
