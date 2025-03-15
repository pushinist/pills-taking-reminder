package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"pills-taking-reminder/internal/service"
)

type Server struct {
	router  *chi.Mux
	service service.StorageService
}

func NewServer(service service.StorageService) *Server {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	return &Server{router: router,
		service: service}
}

func (s *Server) RegisterRoutes() {
	s.router.Post("/schedule", s.postScheduleHandler)
	s.router.Get("/schedules", s.getSchedulesHandler)
	//s.router.Get("/schedule", getScheduleHandler)
	//s.router.Get("/next_takings", getNextTakingsHandler)
}

func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, s.router)
}
