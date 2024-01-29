package test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/server/pb"
	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/storage"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"
	_ "github.com/lib/pq" // PG
	"github.com/pressly/goose"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
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
		os.Getenv("DB_NAME"),
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
	cs.db.Exec(`TRUNCATE event`)
}

func (cs *CalendarSuite) TearDownSuite() {
	defer cs.db.Close()
}

func TestCalendarPost(t *testing.T) {
	suite.Run(t, new(CalendarSuite))
}

func (cs *CalendarSuite) insertTestEvent(ev *storage.Event) uuid.UUID {
	const query = `
		INSERT INTO event (title, date_time, duration, description, user_id, notification_time)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id;
	`
	event := testEvent
	if ev != nil {
		event = ev
	}

	var eventUUID uuid.UUID
	err := cs.db.QueryRow(
		query,
		event.Title,
		event.DateTime,
		event.Duration,
		event.Description,
		event.UserID,
		event.TimeNotification,
	).Scan(&eventUUID)

	cs.Require().NoError(err)
	cs.Require().NotEmpty(eventUUID)

	return eventUUID
}

func (cs *CalendarSuite) TestGetEvent() {
	eventID := cs.insertTestEvent(nil)

	res, err := cs.client.GetEvent(cs.ctx, &pb.EventIdRequest{Id: eventID.String()})
	cs.Require().NoError(err)

	cs.Require().Equal(res.Event.Description, testEvent.Description)
	cs.Require().Equal(res.Event.Title, testEvent.Title)
	cs.Require().Equal(res.Event.UserId, testEvent.UserID)
	cs.Require().Equal(res.Event.Duration, testEvent.Duration)
}

func (cs *CalendarSuite) TestGetEventWithError() {
	_, err := cs.client.GetEvent(cs.ctx, &pb.EventIdRequest{Id: uuid.NewString()})
	cs.Require().Error(err)
}

func (cs *CalendarSuite) TestCreateEvent() {
	req := &pb.EventRequest{
		Event: &pb.Event{
			Title:       "Test Event Title",
			DateTime:    &timestamp.Timestamp{Seconds: time.Now().Unix()},
			Duration:    time.Now().Add(time.Hour).Unix(),
			Description: "Test Description",
			UserId:      123,
		},
	}

	_, err := cs.client.CreateEvent(cs.ctx, req)
	cs.Require().NoError(err)
}

func (cs *CalendarSuite) TestUpdateEvent() {
	eventID := cs.insertTestEvent(nil)

	r, err := cs.client.GetEvent(cs.ctx, &pb.EventIdRequest{Id: eventID.String()})
	cs.Require().NoError(err)

	r.Event.Title = "Update Title"
	req := &pb.EventUpdateRequest{
		Id:    eventID.String(),
		Event: r.Event,
	}

	// update Event
	_, err = cs.client.UpdateEvent(cs.ctx, req)
	cs.Require().NoError(err)

	// get it back and check
	res, err := cs.client.GetEvent(cs.ctx, &pb.EventIdRequest{Id: eventID.String()})
	cs.Require().NoError(err)

	cs.Require().Equal(res.Event.Description, testEvent.Description)
	cs.Require().Equal("Update Title", res.Event.Title)
}

func (cs *CalendarSuite) TestUpdateEventWithError() {
	eventID := cs.insertTestEvent(nil)

	r, err := cs.client.GetEvent(cs.ctx, &pb.EventIdRequest{Id: eventID.String()})
	cs.Require().NoError(err)

	// try to update non-exists Event
	_, err = cs.client.UpdateEvent(cs.ctx, &pb.EventUpdateRequest{Id: "WRONG EVENT UUID", Event: r.Event})
	cs.Require().Error(err)
}

func (cs *CalendarSuite) TestDeleteEvent() {
	eventID := cs.insertTestEvent(nil)

	// check it
	_, err := cs.client.GetEvent(cs.ctx, &pb.EventIdRequest{Id: eventID.String()})
	cs.Require().NoError(err)

	// try to delete them
	_, err = cs.client.DeleteEvent(cs.ctx, &pb.EventIdRequest{Id: eventID.String()})
	cs.Require().NoError(err)

	// there is no Event in databse
	_, err = cs.client.GetEvent(cs.ctx, &pb.EventIdRequest{Id: eventID.String()})
	cs.Require().ErrorContains(err, storage.ErrEventNotFound.Error())
}

func (cs *CalendarSuite) TestDeleteEventWithError() {
	// try to delete unknown Event with UUID
	_, err := cs.client.DeleteEvent(cs.ctx, &pb.EventIdRequest{Id: uuid.NewString()})
	cs.Require().ErrorContains(err, storage.ErrEventNotFound.Error())
}

func (cs *CalendarSuite) TestGetEvents() {
	var eventIds []uuid.UUID

	// create 3 Events
	for i := 0; i < 3; i++ {
		eventIds = append(eventIds, cs.insertTestEvent(nil))
	}

	res, err := cs.client.GetEvents(cs.ctx, &emptypb.Empty{})
	cs.Require().NoError(err)
	cs.Require().Len(res.Events, 3)

	for _, event := range res.Events {
		contains := slices.ContainsFunc(eventIds, func(u uuid.UUID) bool {
			return event.Id == u.String()
		})
		if !contains {
			cs.Require().Failf("ERROR", "There is no '%s' EventId in target slice of Events", event.Id)
		}
	}
}

func (cs *CalendarSuite) TestGetEventsForDay() {
	now := time.Now().UTC()
	ev1 := &storage.Event{
		Title:            "First Event Title",
		DateTime:         now.Add(time.Minute),
		Duration:         now.Add(time.Hour).Unix(),
		Description:      "Test Description",
		UserID:           123,
		TimeNotification: now.Add(time.Hour),
	}
	cs.insertTestEvent(ev1)
	ev2 := &storage.Event{
		Title:            "Second Event Title",
		DateTime:         now.Add(time.Hour),
		Duration:         now.Add(time.Hour * 2).Unix(),
		Description:      "Test Description",
		UserID:           123,
		TimeNotification: now.Add(time.Hour),
	}
	cs.insertTestEvent(ev2)
	ev3 := &storage.Event{
		Title:            "Third Event Title",
		DateTime:         now.Add(time.Hour * 30),
		Duration:         now.Add(time.Hour * 31).Unix(),
		Description:      "Test Description",
		UserID:           123,
		TimeNotification: now.Add(time.Hour),
	}
	cs.insertTestEvent(ev3)

	endOfDay := now.Add(24 * time.Hour).UTC()

	req := &pb.RangeRequest{
		DateTime: &timestamp.Timestamp{Seconds: now.Unix()},
	}
	res, err := cs.client.GetEventsForDay(cs.ctx, req)
	cs.Require().NoError(err)
	cs.Require().Len(res.Events, 2)

	for _, event := range res.Events {
		cs.WithinRange(event.DateTime.AsTime(), now, endOfDay)
	}
}

func (cs *CalendarSuite) TestGetEventsForWeek() {
	now := time.Now().UTC()
	ev1 := &storage.Event{
		Title:            "First Event Title",
		DateTime:         now.Add(time.Hour),
		Duration:         now.Add(time.Hour * 2).Unix(),
		Description:      "Test Description",
		UserID:           123,
		TimeNotification: now.Add(time.Hour),
	}
	cs.insertTestEvent(ev1)
	ev2 := &storage.Event{
		Title:            "Second Event Title",
		DateTime:         now.Add(4 * 24 * time.Hour),
		Duration:         now.Add(8 * 24 * time.Hour).Unix(),
		Description:      "Test Description",
		UserID:           123,
		TimeNotification: now.Add(time.Hour),
	}
	cs.insertTestEvent(ev2)
	ev3 := &storage.Event{
		Title:            "Third Event Title",
		DateTime:         now.Add(10 * 24 * time.Hour),
		Duration:         now.Add(11 * 24 * time.Hour).Unix(),
		Description:      "Test Description",
		UserID:           123,
		TimeNotification: now.Add(time.Hour),
	}
	cs.insertTestEvent(ev3)

	endOfWeek := now.Add(7 * 24 * time.Hour).UTC()

	req := &pb.RangeRequest{
		DateTime: &timestamp.Timestamp{Seconds: now.Unix()},
	}
	res, err := cs.client.GetEventsForWeek(cs.ctx, req)
	cs.Require().NoError(err)
	cs.Require().Len(res.Events, 2)

	for _, event := range res.Events {
		cs.WithinRange(event.DateTime.AsTime(), now, endOfWeek)
	}
}

func (cs *CalendarSuite) TestGetEventsForMonth() {
	now := time.Now().UTC()
	ev1 := &storage.Event{
		Title:            "First Event Title",
		DateTime:         now.Add(time.Hour),
		Duration:         now.Add(time.Hour * 2).Unix(),
		Description:      "Test Description",
		UserID:           123,
		TimeNotification: now.Add(time.Hour),
	}
	cs.insertTestEvent(ev1)
	ev2 := &storage.Event{
		Title:            "Second Event Title",
		DateTime:         now.AddDate(0, 0, 15),
		Duration:         now.AddDate(0, 0, 16).Unix(),
		Description:      "Test Description",
		UserID:           123,
		TimeNotification: now.Add(time.Hour),
	}
	cs.insertTestEvent(ev2)
	ev3 := &storage.Event{
		Title:            "Third Event Title",
		DateTime:         now.AddDate(0, 0, 31),
		Duration:         now.AddDate(0, 0, 32).Unix(),
		Description:      "Test Description",
		UserID:           123,
		TimeNotification: now.Add(time.Hour),
	}
	cs.insertTestEvent(ev3)

	endOfMonth := now.AddDate(0, 1, -1).UTC()

	req := &pb.RangeRequest{
		DateTime: &timestamp.Timestamp{Seconds: now.Unix()},
	}
	res, err := cs.client.GetEventsForMonth(cs.ctx, req)
	cs.Require().NoError(err)
	cs.Require().Len(res.Events, 2)

	for _, event := range res.Events {
		cs.WithinRange(event.DateTime.AsTime(), now, endOfMonth)
	}
}
