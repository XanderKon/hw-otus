package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
	_ "github.com/lib/pq" // PG
	"github.com/pressly/goose"
)

type Storage struct {
	DB               *sql.DB
	connectionString string
	migrationsPath   string
}

func New(connectionString string, migrationsPath string) *Storage {
	return &Storage{
		connectionString: connectionString,
		migrationsPath:   migrationsPath,
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

	return goose.Up(s.DB, s.migrationsPath)
}

func (s *Storage) Close() error {
	return s.DB.Close()
}

func (s *Storage) CreateEvent(ctx context.Context, event *storage.Event) error {
	const query = `
		INSERT INTO event (title, date_time, duration, description, user_id, notification_time)
		VALUES ($1, $2, $3, $4, $5, $6)
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
		SET title = $1, date_time = $2, duration = $3, description = $4, user_id = $5, notification_time = $6, notify_at = $7
		WHERE id = $8
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
		event.NotifyAt,
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

	for rows.Next() {
		e := &storage.Event{}
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

// general mehtod for getting events by date range.
func (s *Storage) getEventsForRange(
	ctx context.Context,
	startRange time.Time,
	endRange time.Time,
) ([]*storage.Event, error) {
	var events []*storage.Event

	const query = `
		SELECT id, title, date_time, duration, description, user_id, notification_time
		FROM event
		WHERE date_time >= $1 AND date_time < $2
	`

	rows, err := s.DB.QueryContext(ctx, query, startRange, endRange)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate on the results of the query and create event objects
	for rows.Next() {
		event := &storage.Event{}
		err := rows.Scan(
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
		events = append(events, event)
	}

	return events, nil
}

func (s *Storage) GetEventsForDay(ctx context.Context, startOfDay time.Time) ([]*storage.Event, error) {
	return s.getEventsForRange(ctx, startOfDay, startOfDay.Add(24*time.Hour))
}

func (s *Storage) GetEventsForWeek(ctx context.Context, startOfWeek time.Time) ([]*storage.Event, error) {
	return s.getEventsForRange(ctx, startOfWeek, startOfWeek.AddDate(0, 0, 7))
}

func (s *Storage) GetEventsForMonth(ctx context.Context, startOfMonth time.Time) ([]*storage.Event, error) {
	return s.getEventsForRange(ctx, startOfMonth, startOfMonth.AddDate(0, 1, 0))
}

func (s *Storage) GetEventsForNotifications(ctx context.Context) ([]*storage.Event, error) {
	var events []*storage.Event

	const query = `
		SELECT id, title, date_time, user_id
		FROM event
		WHERE EXTRACT(EPOCH FROM (notification_time - NOW())) < 0
		AND notify_at is null
	`

	rows, err := s.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate on the results of the query and create event objects
	for rows.Next() {
		event := &storage.Event{}
		err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.DateTime,
			&event.UserID,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

func (s *Storage) DeleteOldEvents(ctx context.Context, duration time.Duration) (int, error) {
	query := fmt.Sprintf("DELETE FROM event WHERE date_time < NOW() - INTERVAL '%d hours'", int(duration.Hours()))

	res, err := s.DB.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return int(affected), err
	}

	return int(affected), nil
}
