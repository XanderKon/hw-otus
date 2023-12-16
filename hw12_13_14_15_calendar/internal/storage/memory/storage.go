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
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, found := s.events[eventID]; !found {
		return storage.ErrEventNotFound
	}

	// busy time
	if _, err := s.GetEventByDate(ctx, event.DateTime); err == nil {
		return storage.ErrEventDateTimeIsBusy
	}

	s.events[eventID] = event

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
	// because can be already locked by parent function.
	if s.mu.TryLock() {
		s.mu.Lock()
		defer s.mu.Unlock()
	}

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

// general mehtod for getting events by date range.
func (s *Storage) getEventsForRange(startRange time.Time, endRange time.Time) ([]*storage.Event, error) {
	var events []*storage.Event
	for _, event := range s.events {
		if event.DateTime.After(startRange.Add(-time.Second)) && event.DateTime.Before(endRange) {
			events = append(events, event)
		}
	}

	return events, nil
}

func (s *Storage) GetEventsForDay(_ context.Context, startOfDay time.Time) ([]*storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.getEventsForRange(startOfDay, startOfDay.Add(24*time.Hour))
}

func (s *Storage) GetEventsForWeek(_ context.Context, startOfWeek time.Time) ([]*storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	endRange := startOfWeek.Add(24 * time.Hour)

	return s.getEventsForRange(startOfWeek, endRange.AddDate(0, 0, 7))
}

func (s *Storage) GetEventsForMonth(_ context.Context, startOfMonth time.Time) ([]*storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	endRange := startOfMonth.Add(24 * time.Hour)

	return s.getEventsForRange(startOfMonth, endRange.AddDate(0, 1, 0))
}

func (s *Storage) GetEventsForNotifications(_ context.Context) ([]*storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var events []*storage.Event
	for _, event := range s.events {
		if time.Until(event.TimeNotification) <= 0 && event.NotifyAt.IsZero() {
			events = append(events, event)
		}
	}

	return events, nil
}
