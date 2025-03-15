package server

import (
	"encoding/json"
	"fmt"
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(id)
	if err != nil {
		log.Printf("error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (s *Server) getSchedulesHandler(w http.ResponseWriter, r *http.Request) {
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
	if len(schedulesIDs) == 0 {
		message := fmt.Sprintf("No schedules found for user with id:%d", userID)
		w.Write([]byte(message))
		w.WriteHeader(http.StatusNotFound)
		return
	}
	err = json.NewEncoder(w).Encode(schedulesIDs)
	if err != nil {
		log.Printf("error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
