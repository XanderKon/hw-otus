package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
	_ "github.com/lib/pq" // PG
	"github.com/pressly/goose"
)

type Storage struct {
	DB               *sql.DB
	connectionString string
}

func New(connectionString string) *Storage {
	return &Storage{
		connectionString: connectionString,
	}
}

func (s *Storage) Connect(ctx context.Context) error {
	db, err := sql.Open("postgres", s.connectionString)
	if err != nil {
		return err
	}

	err = db.PingContext(ctx)
	if err != nil {
		return err
	}

	s.DB = db

	return s.migrate()
}

func (s *Storage) migrate() error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.Up(s.DB, "migrations")
}

func (s *Storage) Close() error {
	return s.DB.Close()
}

func (s *Storage) CreateEvent(ctx context.Context, event *storage.Event) error {
	const query = `
		INSERT INTO event (title, date_time, duration, description, user_id, notification_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := s.DB.ExecContext(
		ctx,
		query,
		event.Title,
		event.DateTime,
		event.Duration,
		event.Description,
		event.UserID,
		event.TimeNotification,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, eventID uuid.UUID, event *storage.Event) error {
	const query = `
		UPDATE event
		SET title = $1, date_time = $2, duration = $3, description = $4, user_id = $5, notification_time = $6
		WHERE id = $7
	`

	_, err := s.DB.ExecContext(
		ctx,
		query,
		event.Title,
		event.DateTime,
		event.Duration,
		event.Description,
		event.UserID,
		event.TimeNotification,
		eventID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, eventID uuid.UUID) error {
	const query = `DELETE FROM event WHERE id = $1`

	_, err := s.DB.ExecContext(ctx, query, eventID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetEvent(ctx context.Context, eventID uuid.UUID) (*storage.Event, error) {
	const query = `
		SELECT id, title, date_time, duration, description, user_id, notification_time
		FROM event
		WHERE id = $1
	`

	row := s.DB.QueryRowContext(ctx, query, eventID)

	if errors.Is(row.Err(), sql.ErrNoRows) {
		return nil, storage.ErrEventNotFound
	}

	var event storage.Event
	err := row.Scan(
		&event.ID,
		&event.Title,
		&event.DateTime,
		&event.Duration,
		&event.Description,
		&event.UserID,
		&event.TimeNotification,
	)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (s *Storage) GetEvents(ctx context.Context) ([]*storage.Event, error) {
	const query = `
		SELECT id, title, date_time, duration, description, user_id, notification_time
		FROM event
	`
	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var events []*storage.Event
	var event storage.Event

	for rows.Next() {
		e := &event
		err := rows.Scan(
			&e.ID,
			&e.Title,
			&e.DateTime,
			&e.Duration,
			&e.Description,
			&e.UserID,
			&e.TimeNotification,
		)
		if err != nil {
			return nil, err
		}

		events = append(events, e)
	}

	return events, nil
}

func (s *Storage) GetEventByDate(ctx context.Context, eventDatetime time.Time) (*storage.Event, error) {
	const query = `
		SELECT id, title, date_time, duration, description, user_id, notification_time
		FROM event
		WHERE date_time = $1
	`

	row := s.DB.QueryRowContext(ctx, query, eventDatetime.String())

	var event storage.Event

	err := row.Scan(
		&event.ID,
		&event.Title,
		&event.DateTime,
		&event.Duration,
		&event.Description,
		&event.UserID,
		&event.TimeNotification,
	)
	if err != nil {
		return nil, err
	}

	return &event, nil
}
