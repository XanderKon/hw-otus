package memorystorage

import (
	"context"
	"sync"

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

func (s *Storage) UpdateEvent(_ context.Context, eventID uuid.UUID, event *storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, found := s.events[eventID]; !found {
		return storage.ErrEventNotFound
	}

	s.events[event.ID] = event
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

func (s *Storage) GetEvents(_ context.Context) ([]*storage.Event, error) {
	return maps.Values(s.events), nil
}
