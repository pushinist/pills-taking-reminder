package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"pills-taking-reminder/internal/models"
)

type StorageService interface {
	CreateSchedule(schedule models.ScheduleRequest) (int64, error)
	GetSchedulesIDs(userID int64) ([]int64, error)
	GetNextTakings(id int64) ([]models.Taking, error)
	GetSchedule(userID, scheduleID int64) (models.ScheduleResponse, error)
}

type Server struct {
	router  *chi.Mux
	service StorageService
}

func NewServer(service StorageService) *Server {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	return &Server{router: router,
		service: service}
}

func (s *Server) RegisterRoutes() {
	s.router.Post("/schedule", s.postScheduleHandler)
	s.router.Get("/schedules", s.getSchedulesIDsHandler)
	s.router.Get("/schedule", s.getScheduleHandler)
	s.router.Get("/next_takings", s.getNextTakingsHandler)
}

func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, s.router)
}
