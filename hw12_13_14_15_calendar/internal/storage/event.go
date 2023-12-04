package storage

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEventAlreadyExists = errors.New("event already exists")
	ErrEventNotFound      = errors.New("event not found")
)

type Event struct {
	ID               uuid.UUID `json:"id"`
	Title            string    `json:"title"`
	DateTime         time.Time `json:"date_time"` //nolint:tagliatelle
	Duration         int64     `json:"duration"`
	Description      string    `json:"description"`
	UserID           string    `json:"user_id"`           //nolint:tagliatelle
	TimeNotification time.Time `json:"time_notification"` //nolint:tagliatelle
}

type Notification struct {
	EventID  string    `json:"event_id"` //nolint:tagliatelle
	Title    string    `json:"title"`
	DateTime time.Time `json:"date_time"` //nolint:tagliatelle
	UserID   string    `json:"user_id"`   //nolint:tagliatelle
}
