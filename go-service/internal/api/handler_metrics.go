package api

import (
	"encoding/json"
	"net/http"

	"github.com/agenticsdk/go-service/internal/metrics"
)

type MetricsHandler struct{}

func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{}
}

func (h *MetricsHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	data := metrics.GlobalTracker.GetMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
