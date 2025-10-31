package http

import (
	"net/http"
	"time"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (handler *HealthHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, GetHealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Unix(),
	})
}
