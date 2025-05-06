package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"pills-taking-reminder/internal/api/dto"
	"pills-taking-reminder/internal/usecase"
	"pills-taking-reminder/pkg/mw"
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
	traceID := mw.GetTraceID(ctx)

	var req dto.ScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body",
			slog.String("error", err.Error()),
			slog.String("trace_id", traceID))
		h.respondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.logger.Error("validation failed",
			slog.String("error", err.Error()),
			slog.String("trace_id", traceID))
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
		h.logger.Error("failed to create schedule",
			slog.String("error", err.Error()),
			slog.String("trace_id", traceID))

		switch {
		case errors.Is(err, usecase.ErrInvalidInput):
			h.respondWithError(w, http.StatusBadRequest, "Invalid input parameters")
		case errors.Is(err, usecase.ErrScheduleExists):
			h.respondWithError(w, http.StatusConflict, "Schedule already exists")
		default:
			h.respondWithJSON(w, http.StatusInternalServerError, "Failed to create schedule")
		}
		return
	}

	h.logger.Info("schedule was created successfully!",
		slog.String("trace_id", traceID))
	h.respondWithJSON(w, http.StatusOK, id)

}

func (h *ScheduleHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	traceID := mw.GetTraceID(ctx)

	userID, err := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		h.logger.Debug("got invalid user id",
			slog.Int64("user_id", userID),
			slog.String("trace_id", traceID))
		h.respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	scheduleID, err := strconv.ParseInt(r.URL.Query().Get("schedule_id"), 10, 64)
	if err != nil || scheduleID <= 0 {
		h.logger.Debug("got invalid schedule id",
			slog.Int64("schedule_id", scheduleID),
			slog.String("trace_id", traceID))
		h.respondWithError(w, http.StatusBadRequest, "Invalid schedule ID")
		return
	}

	schedule, err := h.scheduleUseCase.GetSchedule(ctx, userID, scheduleID)
	if err != nil {
		h.logger.Error("failed to get schedule for user",
			slog.String("error", err.Error()),
			slog.String("trace_id", traceID),
			slog.Int64("user_id", userID),
			slog.Int64("schedule_id", scheduleID))
		switch {
		case errors.Is(err, usecase.ErrScheduleNotFound):
			h.respondWithError(w, http.StatusNotFound, "Schedule was not found")
		case errors.Is(err, usecase.ErrInvalidInput):
			h.respondWithError(w, http.StatusBadRequest, "Invalid input parameters")
		default:
			h.respondWithError(w, http.StatusInternalServerError, "Failed to get schedule")
		}
		return
	}

	response := dto.ScheduleResponse{
		ID:           schedule.ID,
		MedicineName: schedule.MedicineName,
		StartDate:    schedule.StartDate,
		EndDate:      schedule.EndDate,
		UserID:       schedule.UserID,
		TakingTime:   schedule.TakingTimes,
	}

	h.logger.Info("successfully got schedule info",
		slog.String("trace_id", traceID))
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *ScheduleHandler) GetScheduleIDs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	traceID := mw.GetTraceID(ctx)

	userID, err := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		h.logger.Debug("got invalid user id",
			slog.Int64("user_id", userID),
			slog.String("trace_id", traceID))
		h.respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	ids, err := h.scheduleUseCase.GetScheduleIDs(ctx, userID)
	if err != nil {
		h.logger.Error("failed to get schedules IDs for user",
			slog.String("error", err.Error()),
			slog.String("trace_id", traceID),
			slog.Int64("user_id", userID))

		switch {
		case errors.Is(err, usecase.ErrInvalidInput):
			h.respondWithError(w, http.StatusBadRequest, "Invalid request parameters")
		default:
			h.respondWithError(w, http.StatusInternalServerError, "Failed to get schedule IDs")
		}
		return
	}

	h.logger.Info("successfully got schedules IDs",
		slog.String("trace_id", traceID))

	h.respondWithJSON(w, http.StatusOK, ids)
}

func (h *ScheduleHandler) GetNextTakings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	traceID := mw.GetTraceID(ctx)

	userID, err := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		h.logger.Debug("got invalid user id",
			slog.Int64("user_id", userID),
			slog.String("trace_id", traceID))

		h.respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	takings, err := h.scheduleUseCase.GetNextTakings(ctx, userID)
	if err != nil {
		h.logger.Error("failed to get next takings for user",
			slog.String("error", err.Error()),
			slog.String("trace_id", traceID),
			slog.Int64("user_id", userID))
		switch {
		case errors.Is(err, usecase.ErrInvalidInput):
			h.respondWithError(w, http.StatusBadRequest, "Invalid input parameters")
		default:
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

	h.logger.Info("successfully got next takings!",
		slog.String("trace_id", traceID))
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *ScheduleHandler) respondWithJSON(w http.ResponseWriter, code int, payload any) {
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
	h.respondWithJSON(w, code, dto.ErrorResponse{Error: message})
}
