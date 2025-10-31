package log

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
)

type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"

	CtxKeyLogID = "log_id"
)

type Log struct {
	LogID   string `json:"log_id,omitempty"`
	Time    string `json:"time"`
	Level   Level  `json:"level"`
	Message string `json:"message"`
}

func Info(ctx context.Context, message string) {
	m, _ := json.Marshal(Log{
		LogID:   GetLogID(ctx),
		Time:    time.Now().String(),
		Level:   LevelInfo,
		Message: message,
	})
	log.Print(string(m))
}

func Warn(ctx context.Context, message string) {
	m, _ := json.Marshal(Log{
		LogID:   GetLogID(ctx),
		Time:    time.Now().String(),
		Level:   LevelWarn,
		Message: message,
	})
	log.Print(string(m))
}

func Fatal(ctx context.Context, message string) {
	m, _ := json.Marshal(Log{
		LogID:   GetLogID(ctx),
		Time:    time.Now().String(),
		Level:   LevelInfo,
		Message: message,
	})
	log.Fatal(string(m))
}

func Error(ctx context.Context, message string) {
	m, _ := json.Marshal(Log{
		LogID:   GetLogID(ctx),
		Time:    time.Now().String(),
		Level:   LevelError,
		Message: message,
	})
	log.Print(string(m))
}

func InjectNewID(ctx context.Context) (context.Context, string) {
	existingID, _ := ctx.Value(CtxKeyLogID).(string)
	if existingID == "" {
		logID := uuid.NewString()
		ctx = context.WithValue(ctx, CtxKeyLogID, logID)
		return ctx, logID
	}
	return ctx, existingID
}

func GetLogID(ctx context.Context) string {
	val, ok := ctx.Value(CtxKeyLogID).(string)
	if ok {
		return val
	}
	return ""
}
