package httpjson

import (
	"database/sql"
	"log/slog"
	"net/http"
	pprofhttp "net/http/pprof"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/macimizer/credit-decision-service/internal/observability"
	redispkg "github.com/macimizer/credit-decision-service/internal/platform/redispkg"
	"github.com/macimizer/credit-decision-service/internal/service"
)

type RouterParams struct {
	Logger         *slog.Logger
	Metrics        *observability.Metrics
	Clients        *service.ClientService
	Banks          *service.BankService
	Credits        *service.CreditService
	DB             *sql.DB
	Redis          *redispkg.Client
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	ShutdownTimout time.Duration
}

func NewRouter(params RouterParams) http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(30 * time.Second))
	r.Use(LoggingMiddleware(params.Logger))
	r.Use(params.Metrics.HTTPMiddleware)

	health := NewHealthHandler(params.DB, params.Redis)
	r.Get("/healthz", health.Liveness)
	r.Get("/readyz", health.Readiness)
	r.Handle("/metrics", promhttp.Handler())

	r.Mount("/debug/pprof", pprofRoutes())

	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/clients", NewClientHandler(params.Clients).Routes())
		r.Mount("/banks", NewBankHandler(params.Banks).Routes())
		r.Mount("/credits", NewCreditHandler(params.Credits).Routes())
	})

	return r
}

func pprofRoutes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", pprofhttp.Index)
	r.Get("/cmdline", pprofhttp.Cmdline)
	r.Get("/profile", pprofhttp.Profile)
	r.Get("/symbol", pprofhttp.Symbol)
	r.Get("/trace", pprofhttp.Trace)
	r.Get("/allocs", pprofhttp.Handler("allocs").ServeHTTP)
	r.Get("/block", pprofhttp.Handler("block").ServeHTTP)
	r.Get("/goroutine", pprofhttp.Handler("goroutine").ServeHTTP)
	r.Get("/heap", pprofhttp.Handler("heap").ServeHTTP)
	r.Get("/mutex", pprofhttp.Handler("mutex").ServeHTTP)
	r.Get("/threadcreate", pprofhttp.Handler("threadcreate").ServeHTTP)
	return r
}
