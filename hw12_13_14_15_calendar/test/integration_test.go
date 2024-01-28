package test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/server/pb"
	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
	_ "github.com/lib/pq" // PG
	"github.com/pressly/goose"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CalendarSuite struct {
	suite.Suite
	ctx    context.Context
	client pb.CalendarServiceClient
	db     *sql.DB
}

var testEvent = &storage.Event{
	Title:            "Test Event Title",
	DateTime:         time.Now(),
	Duration:         time.Now().Add(time.Hour).Unix(),
	Description:      "Test Description",
	UserID:           123,
	TimeNotification: time.Now().Add(time.Hour),
}

func (cs *CalendarSuite) SetupSuite() {
	cs.ctx = context.Background()

	host, ok := os.LookupEnv("GRPC_HOST")
	if !ok {
		host = "127.0.0.1"
	}

	port, ok := os.LookupEnv("GRPC_PORT")
	if !ok {
		port = "8081"
	}

	grpcConnect, _ := grpc.Dial(host+":"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	cs.client = pb.NewCalendarServiceClient(grpcConnect)

	DSN := fmt.Sprintf(
		"postgres://postgres:postgres@%s:5432/%s?sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_TESTDBNAME"),
	)
	db, err := sql.Open("postgres", DSN)
	cs.NoError(err)

	// run migrations
	mPath, ok := os.LookupEnv("MIGRATIONS_PATH")
	if !ok {
		mPath = "../migrations"
	}

	goose.SetDialect("postgres")
	err = goose.Up(db, mPath)
	cs.NoError(err)

	fmt.Println(DSN)
	fmt.Println(mPath)

	cs.db = db
}

func (cs *CalendarSuite) TearDownTest() {
	cs.db.Exec(`DELETE FROM events`)
}

func (cs *CalendarSuite) TearDownSuite() {
	defer cs.db.Close()
}

func TestCalendarPost(t *testing.T) {
	suite.Run(t, new(CalendarSuite))
}

func (cs *CalendarSuite) insertTestEvent() uuid.UUID {
	const query = `
		INSERT INTO event (title, date_time, duration, description, user_id, notification_time)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id;
	`
	var eventUUID uuid.UUID
	err := cs.db.QueryRow(
		query,
		testEvent.Title,
		testEvent.DateTime,
		testEvent.Duration,
		testEvent.Description,
		testEvent.UserID,
		testEvent.TimeNotification,
	).Scan(&eventUUID)

	cs.Require().NoError(err)
	cs.Require().NotEmpty(eventUUID)

	return eventUUID
}

func (cs *CalendarSuite) TestGetEvent() {
	eventID := cs.insertTestEvent()

	res, err := cs.client.GetEvent(cs.ctx, &pb.EventIdRequest{Id: eventID.String()})
	cs.Require().NoError(err)

	cs.Require().Equal(res.Event.Description, testEvent.Description)
	cs.Require().Equal(res.Event.Title, testEvent.Title)
	cs.Require().Equal(res.Event.UserId, testEvent.UserID)
	cs.Require().Equal(res.Event.Duration, testEvent.Duration)
}
