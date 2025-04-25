package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"pills-taking-reminder/pkg/masking"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

var credentials = map[string]bool{
	"user_id": true,
}

type MaskingHandler struct {
	handler slog.Handler
}

func (h *MaskingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *MaskingHandler) Handle(ctx context.Context, record slog.Record) error {
	var newAttributes []slog.Attr

	record.Attrs(func(attr slog.Attr) bool {
		if credentials[attr.Key] {
			strVal := fmt.Sprint(attr.Value.Any())
			newAttributes = append(newAttributes,
				slog.String(
					attr.Key+"_masked",
					masking.MaskData(attr.Key, strVal),
				))
		} else {
			newAttributes = append(newAttributes,
				attr,
			)
		}
		return true
	})

	newRec := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	for _, attr := range newAttributes {
		newRec.AddAttrs(attr)
	}

	return h.handler.Handle(ctx, newRec)
}

func (h *MaskingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &MaskingHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *MaskingHandler) WithGroup(name string) slog.Handler {
	return &MaskingHandler{handler: h.handler.WithGroup(name)}
}

func SetupLogger(env string) *slog.Logger {
	var handler slog.Handler

	switch env {
	case envLocal:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	case envDev:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	case envProd:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}

	maskingHandler := &MaskingHandler{handler: handler}

	return slog.New(maskingHandler)
}
