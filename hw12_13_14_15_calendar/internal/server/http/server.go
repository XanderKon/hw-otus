package internalhttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/XanderKon/hw-otus/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/exp/slices"
)

var (
	ErrNotEnoughArguments         = errors.New("not enough argumets")
	ErrNotEnoughEventIDArgument   = errors.New("event id argument not found")
	ErrNotEnoughTypeArgument      = errors.New("type argument not found")
	ErrIncorrectTypeArgument      = errors.New("type argument is not valid. Should be: 'day', 'week' or 'month'")
	ErrNotEnoughStartDateArgument = errors.New("start_date argument not found")
	ErrIncorrectStartDateArgument = errors.New("start_date is not valid. Should be datetime string")
	ErrWrongEventUUIDArgument     = errors.New("cannot parse event id argument to UUID")
	ErrIncorrectRequest           = errors.New("incorrect request")
	ErrEventNotFound              = errors.New("event with this UUID is not found")
	ErrServerError                = errors.New("unexpected server error")
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
	GetEventsForDay(ctx context.Context, startOfDay time.Time) ([]*storage.Event, error)
	GetEventsForWeek(ctx context.Context, startOfWeek time.Time) ([]*storage.Event, error)
	GetEventsForMonth(ctx context.Context, startOfMonth time.Time) ([]*storage.Event, error)
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
	// router init
	r := s.initRouter()

	// setup logging middleware
	handlerWitMiddleware := loggingMiddleware(r, s.logger)

	go func() {
		<-ctx.Done()
		s.Stop(ctx)
	}()

	s.server = &http.Server{
		Addr:              fmt.Sprintf("%s:%d", s.host, s.port),
		Handler:           handlerWitMiddleware,
		ReadHeaderTimeout: 20 * time.Second,
	}

	s.logger.Info("http-server is up...")

	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) initRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", s.defaultHandler).Methods(http.MethodGet)
	r.HandleFunc("/event/{id}", s.getEventHandler).Methods(http.MethodGet)
	r.HandleFunc("/event", s.createEventHandler).Methods(http.MethodPost)
	r.HandleFunc("/event/{id}", s.updateEventHandler).Methods(http.MethodPatch, http.MethodPut)
	r.HandleFunc("/event/{id}", s.deleteEventHandler).Methods(http.MethodDelete)
	r.HandleFunc("/event", s.getAllEventsHandler).Methods(http.MethodGet)

	return r
}

func (s *Server) defaultHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World!"))
}

func (s *Server) getEventHandler(w http.ResponseWriter, r *http.Request) {
	eventUUID, err := s.parseRequestAndGetUUID(r)
	if err != nil {
		s.errorResponse(w, err, http.StatusBadRequest)
		return
	}

	event, err := s.app.GetEvent(r.Context(), eventUUID)
	if err != nil {
		s.errorResponse(w, ErrEventNotFound, http.StatusNotFound)
		return
	}

	s.jsonResponse(w, event)
}

func (s *Server) createEventHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var event storage.Event
	if err := decoder.Decode(&event); err != nil {
		s.errorResponse(w, ErrIncorrectRequest, http.StatusBadRequest)
		return
	}

	err := s.app.CreateEvent(r.Context(), &event)
	if err != nil {
		s.errorResponse(w, ErrServerError, http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, event)
}

func (s *Server) updateEventHandler(w http.ResponseWriter, r *http.Request) {
	eventUUID, err := s.parseRequestAndGetUUID(r)
	if err != nil {
		s.errorResponse(w, err, http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var eventForUpdate storage.Event

	if err := decoder.Decode(&eventForUpdate); err != nil {
		s.errorResponse(w, ErrIncorrectRequest, http.StatusBadRequest)
		return
	}

	if err := s.app.UpdateEvent(r.Context(), eventUUID, &eventForUpdate); err != nil {
		switch {
		case errors.Is(err, storage.ErrEventNotFound):
			s.errorResponse(w, err, http.StatusNotFound)
		case errors.Is(err, storage.ErrEventDateTimeIsBusy):
			s.errorResponse(w, err, http.StatusConflict)
		default:
			s.errorResponse(w, ErrServerError, http.StatusInternalServerError)
		}

		return
	}
}

func (s *Server) deleteEventHandler(w http.ResponseWriter, r *http.Request) {
	eventUUID, err := s.parseRequestAndGetUUID(r)
	if err != nil {
		s.errorResponse(w, err, http.StatusBadRequest)
		return
	}

	// try to get event by UUID
	_, err = s.app.GetEvent(r.Context(), eventUUID)
	if err != nil {
		s.errorResponse(w, ErrEventNotFound, http.StatusNotFound)
		return
	}

	if err := s.app.DeleteEvent(r.Context(), eventUUID); err != nil {
		s.errorResponse(w, ErrEventNotFound, http.StatusNotFound)
		return
	}

	s.jsonResponse(w, "")
}

func (s *Server) getAllEventsHandler(w http.ResponseWriter, r *http.Request) {
	reqType := r.FormValue("type")

	if reqType != "" {
		availableTypes := []string{"day", "week", "month"}
		if !slices.Contains(availableTypes, reqType) {
			s.errorResponse(w, ErrIncorrectTypeArgument, http.StatusBadRequest)
			return
		}
	}

	startDate := r.FormValue("start_date")
	if startDate == "" {
		startDate = time.Now().Format("2006-01-02")
	}

	parsedDate, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		s.errorResponse(w, ErrIncorrectStartDateArgument, http.StatusBadRequest)
		return
	}

	var events []*storage.Event
	switch {
	case reqType == "day":
		events, err = s.app.GetEventsForDay(r.Context(), parsedDate)
		if err != nil {
			s.errorResponse(w, ErrServerError, http.StatusInternalServerError)
			return
		}
	case reqType == "week":
		events, err = s.app.GetEventsForWeek(r.Context(), parsedDate)
		if err != nil {
			s.errorResponse(w, ErrServerError, http.StatusInternalServerError)
			return
		}
	case reqType == "month":
		events, err = s.app.GetEventsForMonth(r.Context(), parsedDate)
		if err != nil {
			s.errorResponse(w, ErrServerError, http.StatusInternalServerError)
			return
		}
	default:
		events, err = s.app.GetEvents(r.Context())
		if err != nil {
			s.errorResponse(w, ErrServerError, http.StatusInternalServerError)
			return
		}
	}

	s.jsonResponse(w, events)
}

// helper for getting event UUID from request.
func (s *Server) parseRequestAndGetUUID(r *http.Request) (uuid.UUID, error) {
	vars := mux.Vars(r)
	eventID, ok := vars["id"]

	if !ok {
		s.logger.Debug(ErrNotEnoughEventIDArgument.Error())
		return uuid.Nil, ErrNotEnoughEventIDArgument
	}

	eventUUID, err := uuid.Parse(eventID)
	if err != nil {
		s.logger.Debug(ErrWrongEventUUIDArgument.Error())
		return uuid.Nil, ErrWrongEventUUIDArgument
	}

	return eventUUID, nil
}

func (s *Server) errorResponse(w http.ResponseWriter, err error, status int) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Response{"error", status, nil, err.Error()})
}

func (s *Server) jsonResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Response{"success", http.StatusOK, data, ""})
}
