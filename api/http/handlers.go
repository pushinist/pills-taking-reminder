package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"pills-taking-reminder/api/dto"
	"pills-taking-reminder/internal/usecase"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type ScheduleHandler struct {
	scheduleUseCase *usecase.ScheduleUseCase
	logger          *slog.Logger
	validate        *validator.Validate
}

func NewScheduleHandler(useCase *usecase.ScheduleUseCase, logger *slog.Logger) *ScheduleHandler {
	return &ScheduleHandler{
		scheduleUseCase: useCase,
		logger:          logger,
		validate:        validator.New(),
	}
}

func (h *ScheduleHandler) RegisterRoutes(r chi.Router) {
	r.Post("/schedule", h.CreateSchedule)
	r.Get("/schedules", h.GetScheduleIDs)
	r.Get("/schedule", h.GetSchedule)
	r.Get("/next_takings", h.GetNextTakings)
}

func (h *ScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req dto.ScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", slog.String("error", err.Error()))
		h.respondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.logger.Error("validation failed", slog.String("error", err.Error()))
		h.respondWithError(w, http.StatusBadRequest, "Invalid request parameters")
		return
	}
	input := usecase.ScheduleInput{
		MedicineName: req.MedicineName,
		Frequency:    req.Frequency,
		Duration:     req.Duration,
		UserID:       req.UserID,
	}

	id, err := h.scheduleUseCase.CreateSchedule(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidInput):
			h.respondWithError(w, http.StatusBadRequest, "Invalid input parameters")
		case errors.Is(err, usecase.ErrScheduleExists):
			h.respondWithError(w, http.StatusConflict, "Schedule already exists")
		default:
			h.logger.Error("failed to create schedule", slog.String("error", err.Error()))
			h.reposndWithJSON(w, http.StatusInternalServerError, "Failed to create schedule")
		}
		return
	}

	h.reposndWithJSON(w, http.StatusOK, id)

}

func (h *ScheduleHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		h.respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	scheduleID, err := strconv.ParseInt(r.URL.Query().Get("schedule_id"), 10, 64)
	if err != nil || scheduleID <= 0 {
		h.respondWithError(w, http.StatusBadRequest, "Invalid schedule ID")
		return
	}

	schedule, err := h.scheduleUseCase.GetSchedule(ctx, userID, scheduleID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrScheduleNotFound):
			h.respondWithError(w, http.StatusNotFound, "Schedule was not found")
		case errors.Is(err, usecase.ErrInvalidInput):
			h.respondWithError(w, http.StatusBadRequest, "Invalid input parameters")
		default:
			h.logger.Error("failed to get schedule", slog.String("error", err.Error()))
			h.respondWithError(w, http.StatusInternalServerError, "Failed to get schedule")
		}
		return
	}

	reponse := dto.ScheduleResponse{
		ID:           schedule.ID,
		MedicineName: schedule.MedicineName,
		StartDate:    schedule.StartDate,
		EndDate:      schedule.EndDate,
		UserID:       schedule.UserID,
		TakingTime:   schedule.TakingTimes,
	}

	h.reposndWithJSON(w, http.StatusOK, reponse)
}

func (h *ScheduleHandler) GetScheduleIDs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		h.respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	ids, err := h.scheduleUseCase.GetScheduleIDs(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidInput):
			h.respondWithError(w, http.StatusBadRequest, "Invalid request parameters")
		default:
			h.logger.Error("failed to get schedule IDs", slog.String("error", err.Error()))
			h.respondWithError(w, http.StatusInternalServerError, "Failed to get schedule IDs")
		}
		return
	}

	h.reposndWithJSON(w, http.StatusOK, ids)
}

func (h *ScheduleHandler) GetNextTakings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		h.respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	takings, err := h.scheduleUseCase.GetNextTakings(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidInput):
			h.respondWithError(w, http.StatusBadRequest, "Invalid input parameters")
		default:
			h.logger.Error("failed to get next takings", slog.String("error", err.Error()))
			h.respondWithError(w, http.StatusInternalServerError, "Failed to get next takings")
		}
		return
	}

	response := make([]dto.Taking, len(takings))
	for i, taking := range takings {
		response[i] = dto.Taking{
			MedicineName: taking.MedicineName,
			TakingTime:   taking.TakingTime,
		}
	}

	h.reposndWithJSON(w, http.StatusOK, response)
}

func (h *ScheduleHandler) reposndWithJSON(w http.ResponseWriter, code int, payload any) {
	response, err := json.Marshal(payload)
	if err != nil {
		h.logger.Error("failed to marshal response", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (h *ScheduleHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.reposndWithJSON(w, code, dto.ErrorResponse{Error: message})
}
