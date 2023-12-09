package internalhttp

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
)

type Server struct {
	host   string
	port   int
	logger Logger
	app    Application
	server *http.Server
}

type Logger interface {
	Debug(msg string, a ...any)
	Info(msg string, a ...any)
	Warning(msg string, a ...any)
	Error(msg string, a ...any)
}

type Application interface {
	CreateEvent(ctx context.Context, event *storage.Event) error
	UpdateEvent(ctx context.Context, eventID uuid.UUID, event *storage.Event) error
	DeleteEvent(ctx context.Context, eventID uuid.UUID) error
	GetEvents(ctx context.Context) ([]*storage.Event, error)
	GetEvent(ctx context.Context, eventID uuid.UUID) (*storage.Event, error)
	GetEventByDate(ctx context.Context, eventDatetime time.Time) (*storage.Event, error)
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
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.defaultHandler)

	// setup middleware for logging
	handlerWitMiddleware := loggingMiddleware(mux, s.logger)

	go func() {
		<-ctx.Done()
		s.Stop(ctx)
	}()

	s.server = &http.Server{
		Addr:              fmt.Sprintf("%s:%d", s.host, s.port),
		Handler:           handlerWitMiddleware,
		ReadHeaderTimeout: 20 * time.Second,
	}

	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) defaultHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World!"))
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, req)
		log.Printf("%s %s %s", req.Method, req.RequestURI, time.Since(start))
	})
}
