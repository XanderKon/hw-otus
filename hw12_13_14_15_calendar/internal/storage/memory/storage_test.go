package memorystorage

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateAndGetAndUpdateEvent(t *testing.T) {
	st := New()

	event := &storage.Event{
		ID:    uuid.New(),
		Title: "Event description",
	}

	err := st.CreateEvent(context.Background(), event)
	assert.NoError(t, err)

	// error if already exists
	err = st.CreateEvent(context.Background(), event)
	assert.Equal(t, storage.ErrEventAlreadyExists, err)

	// update
	event.Title = "Event after update"
	err = st.UpdateEvent(context.Background(), event.ID, event)
	assert.NoError(t, err)

	// check after update
	updatedEvent, err := st.GetEvent(context.Background(), event.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Event after update", updatedEvent.Title)

	// update event that doesn't exist
	err = st.UpdateEvent(context.Background(), uuid.New(), &storage.Event{Title: "1"})
	assert.Equal(t, storage.ErrEventNotFound, err)

	// get event that doesn't exist
	_, err = st.GetEvent(context.Background(), uuid.New())
	assert.Equal(t, storage.ErrEventNotFound, err)
}

func TestDeleteEvent(t *testing.T) {
	st := New()
	event := &storage.Event{
		ID:       uuid.New(),
		Title:    "Event Title",
		DateTime: time.Now(),
	}

	// Create an event
	err := st.CreateEvent(context.Background(), event)
	assert.NoError(t, err)

	// Delete the event
	err = st.DeleteEvent(context.Background(), event.ID)
	assert.NoError(t, err)

	// Try deleting a non-existing event
	err = st.DeleteEvent(context.Background(), uuid.New())
	assert.Equal(t, storage.ErrEventNotFound, err)
}

func TestConcurrent(t *testing.T) {
	st := New()
	UUID := uuid.New()

	event := &storage.Event{
		ID:    UUID,
		Title: "Event Title",
	}
	err := st.CreateEvent(context.Background(), event)
	assert.NoError(t, err)

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			event := &storage.Event{
				ID:    UUID,
				Title: uuid.New().String(),
			}
			err := st.UpdateEvent(context.Background(), UUID, event)
			assert.NoError(t, err)
		}()
	}

	wg.Wait()

	updatedEvent, err := st.GetEvent(context.Background(), UUID)
	fmt.Println(updatedEvent)
	assert.NoError(t, err)
	assert.NotNil(t, updatedEvent)
	assert.NotContains(t, updatedEvent.Title, "Event Title")

	errCh := make(chan error, 50)
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := st.DeleteEvent(context.Background(), UUID)
			if err != nil {
				errCh <- err
			}
		}()
	}

	wg.Wait()
	close(errCh)

	// count how many errors do we have. 49 -- because we delete exactly one
	assert.Equal(t, 49, len(errCh))
}
