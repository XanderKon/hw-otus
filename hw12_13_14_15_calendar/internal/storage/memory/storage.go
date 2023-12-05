package memorystorage

import (
	"context"
	"sync"
	"time"

	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
	"golang.org/x/exp/maps"
)

type Storage struct {
	mu     sync.RWMutex
	events map[uuid.UUID]*storage.Event
}

func New() *Storage {
	return &Storage{
		events: make(map[uuid.UUID]*storage.Event),
	}
}

func (s *Storage) Connect(_ context.Context) error {
	return nil
}

func (s *Storage) Close() error {
	return nil
}

func (s *Storage) CreateEvent(_ context.Context, event *storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, found := s.events[event.ID]; found {
		return storage.ErrEventAlreadyExists
	}

	s.events[event.ID] = event
	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, eventID uuid.UUID, event *storage.Event) error {
	// same id
	if _, found := s.events[eventID]; !found {
		return storage.ErrEventNotFound
	}

	// busy time
	if _, err := s.GetEventByDate(ctx, event.DateTime); err == nil {
		return storage.ErrEventDateTimeIsBusy
	}

	s.mu.Lock()
	s.events[eventID] = event
	s.mu.Unlock()

	return nil
}

func (s *Storage) DeleteEvent(_ context.Context, eventID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, found := s.events[eventID]; !found {
		return storage.ErrEventNotFound
	}

	delete(s.events, eventID)
	return nil
}

func (s *Storage) GetEvent(_ context.Context, eventID uuid.UUID) (*storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	event, found := s.events[eventID]
	if !found {
		return nil, storage.ErrEventNotFound
	}

	return event, nil
}

func (s *Storage) GetEventByDate(_ context.Context, eventDatetime time.Time) (*storage.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, event := range s.events {
		if event.DateTime == eventDatetime {
			return event, nil
		}
	}

	return nil, storage.ErrEventNotFound
}

func (s *Storage) GetEvents(_ context.Context) ([]*storage.Event, error) {
	return maps.Values(s.events), nil
}
