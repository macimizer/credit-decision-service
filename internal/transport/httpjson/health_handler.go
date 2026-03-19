package httpjson

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	redispkg "github.com/macimizer/credit-decision-service/internal/platform/redispkg"
)

type HealthHandler struct {
	db    *sql.DB
	redis *redispkg.Client
}

func NewHealthHandler(db *sql.DB, redis *redispkg.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

func (h *HealthHandler) Liveness(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: err.Error()})
		return
	}
	if err := h.redis.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
