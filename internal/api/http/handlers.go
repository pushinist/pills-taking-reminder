package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"pills-taking-reminder/internal/api/dto"
	api "pills-taking-reminder/internal/api/http/generated"
	"pills-taking-reminder/internal/domain/usecase"
	"pills-taking-reminder/pkg/mw"

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
	handler := api.HandlerWithOptions(h, api.ChiServerOptions{
		BaseRouter: r,
	})
	r.Mount("/", handler)
}

func (h *ScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	traceID := mw.GetTraceID(ctx)

	var req api.CreateScheduleJSONRequestBody
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
	duration := 0
	if req.Duration != nil {
		duration = *req.Duration
	}

	input := usecase.ScheduleInput{
		MedicineName: req.MedicineName,
		Frequency:    req.Frequency,
		Duration:     duration,
		UserID:       req.UserId,
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

func (h *ScheduleHandler) GetSchedule(w http.ResponseWriter, r *http.Request, params api.GetScheduleParams) {
	ctx := r.Context()
	traceID := mw.GetTraceID(ctx)

	schedule, err := h.scheduleUseCase.GetSchedule(ctx, params.UserId, params.ScheduleId)
	if err != nil {
		h.logger.Error("failed to get schedule for user",
			slog.String("error", err.Error()),
			slog.String("trace_id", traceID),
			slog.Int64("user_id", params.UserId),
			slog.Int64("schedule_id", params.ScheduleId))
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

	response := api.ScheduleResponse{
		Id:           &schedule.ID,
		MedicineName: &schedule.MedicineName,
		StartDate:    &schedule.StartDate,
		EndDate:      &schedule.EndDate,
		UserId:       &schedule.UserID,
		TakingTime:   &schedule.TakingTimes,
	}

	h.logger.Info("successfully got schedule info",
		slog.String("trace_id", traceID))
	h.respondWithJSON(w, http.StatusOK, response)
}

func (h *ScheduleHandler) GetScheduleIDs(w http.ResponseWriter, r *http.Request, params api.GetScheduleIDsParams) {
	ctx := r.Context()
	traceID := mw.GetTraceID(ctx)

	ids, err := h.scheduleUseCase.GetScheduleIDs(ctx, params.UserId)
	if err != nil {
		h.logger.Error("failed to get schedules IDs for user",
			slog.String("error", err.Error()),
			slog.String("trace_id", traceID),
			slog.Int64("user_id", params.UserId))

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

func (h *ScheduleHandler) GetNextTakings(w http.ResponseWriter, r *http.Request, params api.GetNextTakingsParams) {
	ctx := r.Context()
	traceID := mw.GetTraceID(ctx)

	takings, err := h.scheduleUseCase.GetNextTakings(ctx, params.UserId)
	if err != nil {
		h.logger.Error("failed to get next takings for user",
			slog.String("error", err.Error()),
			slog.String("trace_id", traceID),
			slog.Int64("user_id", params.UserId))
		switch {
		case errors.Is(err, usecase.ErrInvalidInput):
			h.respondWithError(w, http.StatusBadRequest, "Invalid input parameters")
		default:
			h.respondWithError(w, http.StatusInternalServerError, "Failed to get next takings")
		}
		return
	}

	response := make([]api.Taking, len(takings))
	for i, taking := range takings {
		response[i] = api.Taking{
			MedicineName: &taking.MedicineName,
			TakingTime:   &taking.TakingTime,
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
	if _, err := w.Write(response); err != nil {
		h.logger.Error("failed to write response into writer", slog.String("error", err.Error()))
	}
}

func (h *ScheduleHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, dto.ErrorResponse{Error: message})
}
