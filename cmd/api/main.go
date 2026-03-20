package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/macimizer/credit-decision-service/internal/config"
	"github.com/macimizer/credit-decision-service/internal/infrastructure/events"
	postgresrepo "github.com/macimizer/credit-decision-service/internal/infrastructure/repository/postgres"
	"github.com/macimizer/credit-decision-service/internal/observability"
	postgresplatform "github.com/macimizer/credit-decision-service/internal/platform/postgres"
	redispkg "github.com/macimizer/credit-decision-service/internal/platform/redispkg"
	"github.com/macimizer/credit-decision-service/internal/service"
	httpjson "github.com/macimizer/credit-decision-service/internal/transport/httpjson"
)

func main() {
	cfg := config.Load()
	logger := observability.NewLogger(cfg.LogLevel)
	metrics := observability.NewMetrics()

	ctx := context.Background()
	db, err := postgresplatform.Open(cfg.PostgresDSN)
	if err != nil {
		logger.Error("failed to connect to postgres", slog.Any("error", err))
		os.Exit(1)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			logger.Warn("failed to close postgres connection", slog.Any("error", closeErr))
		}
	}()

	if cfg.AutoMigrate {
		migrateCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := postgresplatform.Migrate(migrateCtx, db); err != nil {
			logger.Error("failed to run migrations", slog.Any("error", err))
			os.Exit(1)
		}
	}

	redisClient := redispkg.New(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	defer func() {
		_ = redisClient.Native().Close()
	}()

	clientRepo := postgresrepo.NewClientRepository(db)
	bankRepo := postgresrepo.NewBankRepository(db)
	creditRepo := postgresrepo.NewCreditRepository(db)
	eventPublisher := events.NewRedisStreamPublisher(redisClient.Native(), cfg.EventStreamName)

	clientService := service.NewClientService(clientRepo, redisClient, cfg.CacheTTL, logger)
	bankService := service.NewBankService(bankRepo, redisClient, cfg.CacheTTL, logger)
	creditService := service.NewCreditService(
		creditRepo,
		clientRepo,
		bankRepo,
		eventPublisher,
		service.NewRuleBasedDecisionEngine(),
		cfg.WorkerCount,
		logger,
		metrics,
	)

	router := httpjson.NewRouter(httpjson.RouterParams{
		Logger:       logger,
		Metrics:      metrics,
		Clients:      clientService,
		Banks:        bankService,
		Credits:      creditService,
		DB:           db,
		Redis:        redisClient,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       60 * time.Second,
	}

	shutdownErrors := make(chan error, 1)
	go func() {
		logger.Info("starting api server", slog.String("addr", cfg.HTTPAddr), slog.String("service", cfg.AppName))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			shutdownErrors <- fmt.Errorf("listen and serve: %w", err)
		}
	}()

	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-shutdownErrors:
		logger.Error("server failed", slog.Any("error", err))
		os.Exit(1)
	case <-signalCtx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Info("server stopped")
}
