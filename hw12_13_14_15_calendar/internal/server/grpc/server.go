package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/server/pb"
	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var ErrWrongEventUUIDArgument = errors.New("cannot parse event id argument to UUID")

type Server struct {
	host   string
	port   int
	logger Logger
	app    Application
	server *grpc.Server
	pb.UnimplementedCalendarServiceServer
}

type Application interface {
	CreateEvent(ctx context.Context, event *storage.Event) error
	UpdateEvent(ctx context.Context, eventID uuid.UUID, event *storage.Event) error
	DeleteEvent(ctx context.Context, eventID uuid.UUID) error
	GetEvents(ctx context.Context) ([]*storage.Event, error)
	GetEvent(ctx context.Context, eventID uuid.UUID) (*storage.Event, error)
	GetEventByDate(ctx context.Context, eventDatetime time.Time) (*storage.Event, error)
	GetEventsForDay(ctx context.Context, startOfDay time.Time) ([]*storage.Event, error)
	GetEventsForWeek(ctx context.Context, startOfWeek time.Time) ([]*storage.Event, error)
	GetEventsForMonth(ctx context.Context, startOfMonth time.Time) ([]*storage.Event, error)
}

type Logger interface {
	Debug(msg string, a ...any)
	Info(msg string, a ...any)
	Warning(msg string, a ...any)
	Error(msg string, a ...any)
}

type Response struct {
	Status     string `json:"status"`
	StatusCode int    `json:"statusCode"`
	Data       any    `json:"data"`
	Error      string `json:"error"`
}

func NewServer(host string, port int, logger Logger, app Application) *Server {
	return &Server{
		host:   host,
		port:   port,
		logger: logger,
		app:    app,
	}
}

func (s *Server) Start(ctx context.Context) error {
	lsn, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		return err
	}

	// init interceptor.
	s.server = grpc.NewServer(
		grpc.UnaryInterceptor(NewLoggingInterceptor(s.logger).UnaryServerLoggingInterceptor),
	)
	pb.RegisterCalendarServiceServer(s.server, s)

	// init reflection/
	reflection.Register(s.server)

	go func() {
		<-ctx.Done()
		s.Stop()
	}()

	s.logger.Info("grpc-server is up...")

	return s.server.Serve(lsn)
}

func (s *Server) Stop() {
	s.server.GracefulStop()
}

func (s *Server) CreateEvent(ctx context.Context, req *pb.EventRequest) (*emptypb.Empty, error) {
	event := &storage.Event{
		Title:            req.Event.Title,
		Description:      req.Event.Description,
		UserID:           req.Event.UserId,
		Duration:         req.Event.Duration,
		TimeNotification: req.Event.TimeNotification.AsTime(),
		DateTime:         req.Event.DateTime.AsTime(),
	}

	err := s.app.CreateEvent(ctx, event)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) UpdateEvent(ctx context.Context, req *pb.EventUpdateRequest) (*emptypb.Empty, error) {
	event := &storage.Event{
		Title:            req.Event.Title,
		Description:      req.Event.Description,
		UserID:           req.Event.UserId,
		Duration:         req.Event.Duration,
		TimeNotification: req.Event.TimeNotification.AsTime(),
		DateTime:         req.Event.DateTime.AsTime(),
	}

	eventUUID, err := s.parseRequestAndGetUUID(req.Id)
	if err != nil {
		return nil, err
	}

	err = s.app.UpdateEvent(ctx, eventUUID, event)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) GetEvents(ctx context.Context, _ *emptypb.Empty) (*pb.EventsResponse, error) {
	events, err := s.app.GetEvents(ctx)
	if err != nil {
		return nil, err
	}

	pbEvents := make([]*pb.Event, len(events))
	for i, event := range events {
		pbEvents[i] = &pb.Event{
			Id:               event.ID.String(),
			Title:            event.Title,
			Description:      event.Description,
			UserId:           event.UserID,
			Duration:         event.Duration,
			TimeNotification: timestamppb.New(event.TimeNotification),
			DateTime:         timestamppb.New(event.DateTime),
		}
	}

	eventResponse := &pb.EventsResponse{
		Events: pbEvents,
	}
	return eventResponse, nil
}

func (s *Server) DeleteEvent(ctx context.Context, req *pb.EventIdRequest) (*emptypb.Empty, error) {
	eventUUID, err := s.parseRequestAndGetUUID(req.Id)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	err = s.app.DeleteEvent(ctx, eventUUID)
	if err != nil {
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) GetEvent(ctx context.Context, req *pb.EventIdRequest) (*pb.EventResponse, error) {
	eventUUID, err := s.parseRequestAndGetUUID(req.Id)
	if err != nil {
		return nil, err
	}

	event, err := s.app.GetEvent(ctx, eventUUID)
	if err != nil {
		return nil, err
	}

	return s.eventResponse(event), nil
}

func (s *Server) GetEventsForDay(ctx context.Context, req *pb.RangeRequest) (*pb.EventsResponse, error) {
	fmt.Println(req.DateTime)
	fmt.Println(req.DateTime.AsTime())

	events, err := s.app.GetEventsForDay(ctx, req.DateTime.AsTime())
	if err != nil {
		return nil, err
	}

	return &pb.EventsResponse{Events: s.eventsReponse(events)}, nil
}

func (s *Server) GetEventsForWeek(ctx context.Context, req *pb.RangeRequest) (*pb.EventsResponse, error) {
	events, err := s.app.GetEventsForWeek(ctx, req.DateTime.AsTime())
	if err != nil {
		return nil, err
	}

	return &pb.EventsResponse{Events: s.eventsReponse(events)}, nil
}

func (s *Server) GetEventsForMonth(ctx context.Context, req *pb.RangeRequest) (*pb.EventsResponse, error) {
	events, err := s.app.GetEventsForMonth(ctx, req.DateTime.AsTime())
	if err != nil {
		return nil, err
	}

	return &pb.EventsResponse{Events: s.eventsReponse(events)}, nil
}

// helper for getting event UUID from request.
func (s *Server) parseRequestAndGetUUID(uuidString string) (uuid.UUID, error) {
	eventUUID, err := uuid.Parse(uuidString)
	if err != nil {
		s.logger.Debug(ErrWrongEventUUIDArgument.Error())
		return uuid.Nil, ErrWrongEventUUIDArgument
	}

	return eventUUID, nil
}

func (s *Server) eventResponse(event *storage.Event) *pb.EventResponse {
	return &pb.EventResponse{
		Event: &pb.Event{
			Id:               event.ID.String(),
			Title:            event.Title,
			Description:      event.Description,
			UserId:           event.UserID,
			Duration:         event.Duration,
			TimeNotification: timestamppb.New(event.TimeNotification),
			DateTime:         timestamppb.New(event.DateTime),
		},
	}
}

func (s *Server) eventsReponse(events []*storage.Event) []*pb.Event {
	res := make([]*pb.Event, len(events))
	for i, event := range events {
		res[i] = &pb.Event{
			Id:               event.ID.String(),
			Title:            event.Title,
			Description:      event.Description,
			UserId:           event.UserID,
			Duration:         event.Duration,
			TimeNotification: timestamppb.New(event.TimeNotification),
			DateTime:         timestamppb.New(event.DateTime),
		}
	}
	return res
}
