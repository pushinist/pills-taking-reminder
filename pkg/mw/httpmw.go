package mw

import (
	"context"
	"log/slog"
	"maps"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type contextKey string

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	data       int64
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	if err != nil {
		slog.Error("error writing data to writer",
			slog.String("error", err.Error()))
	}
	w.data += int64(n)
	return n, err
}

func HTTPLoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := r.Header.Get("X-TRACE-ID")
			if traceID == "" {
				traceID = uuid.New().String()
			}
			clientIP := r.RemoteAddr
			userAgent := r.Header.Get("User-Agent")

			w.Header().Set("X-TRACE-ID", traceID)

			startTime := time.Now()

			logger.Info("got request in http",
				slog.String("url", r.URL.String()),
				slog.String("path", r.URL.Path),
				slog.String("method", r.Method),
				slog.String("trace_id", traceID),
				slog.String("client_ip", clientIP),
				slog.String("user_agent", userAgent),
				slog.Any("header", getHeaders(r)),
				slog.Any("params", getParams(r)),
				slog.Int64("timestamp", startTime.Unix()))

			ctx := context.WithValue(r.Context(), contextKey("trace_id"), traceID)
			ctx = context.WithValue(ctx, contextKey("client_ip"), clientIP)
			ctx = context.WithValue(ctx, contextKey("user_agent"), userAgent)

			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(rw, r.WithContext(ctx))

			duration := time.Since(startTime)
			endTime := time.Now()

			logger.Info("request complete in http",
				slog.String("url", r.URL.String()),
				slog.String("method", r.Method),
				slog.String("trace_id", traceID),
				slog.String("client_ip", clientIP),
				slog.String("user_agent", userAgent),
				slog.Int("status_code", rw.statusCode),
				slog.Duration("process_time", duration),
				slog.Int64("response_size", rw.data),
				slog.Int64("timestamp", endTime.Unix()))

		})
	}
}

func getHeaders(r *http.Request) map[string][]string {
	headers := make(map[string][]string)
	for name, values := range r.Header {
		if name == "Authorization" || name == "Cookie" {
			headers[name] = []string{"[SENSITIVE DATA]"}
			continue
		}
		headers[name] = values
	}
	return headers
}

func getParams(r *http.Request) map[string][]string {
	params := make(map[string][]string)
	maps.Copy(params, r.URL.Query())
	return params
}

func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(contextKey("trace_id")).(string); ok {
		return traceID
	}
	return ""
}
