package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"log"
	"net/http"
	"pills-taking-reminder/internal/models"
	"strconv"
)

func (s *Server) postScheduleHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var schedule models.ScheduleRequest
	err := decoder.Decode(&schedule)
	if err != nil {
		log.Printf("error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	validate := validator.New()
	if err := validate.Struct(schedule); err != nil {
		log.Printf("error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := s.service.CreateSchedule(schedule)
	if err != nil {
		log.Printf("error: %v", err)
		w.WriteHeader(http.StatusConflict)
		return
	}

	err = json.NewEncoder(w).Encode(id)
	if err != nil {
		log.Printf("error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (s *Server) getSchedulesIDsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
	if err != nil {
		log.Printf("error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	schedulesIDs, err := s.service.GetSchedulesIDs(userID)
	if err != nil {
		log.Printf("error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(schedulesIDs)
	if err != nil {
		log.Printf("error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (s *Server) getScheduleHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
	if err != nil {
		log.Printf("error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	scheduleID, err := strconv.ParseInt(r.URL.Query().Get("schedule_id"), 10, 64)
	if err != nil {
		log.Printf("error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	schedule, err := s.service.GetSchedule(userID, scheduleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
		}
		log.Printf("error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.NewEncoder(w).Encode(schedule)
	if err != nil {
		log.Printf("error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

}

func (s *Server) getNextTakingsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
	if err != nil {
		log.Printf("error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	nextTakings, err := s.service.GetNextTakings(userID)
	if err != nil {
		log.Printf("error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(nextTakings)
	if err != nil {
		log.Printf("error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
