package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/macimizer/credit-decision-service/internal/domain"
	"github.com/macimizer/credit-decision-service/internal/observability"
	"github.com/macimizer/credit-decision-service/internal/ports"
)

type CreditService struct {
	creditRepo   ports.CreditRepository
	clientRepo   ports.ClientRepository
	bankRepo     ports.BankRepository
	eventBus     ports.EventPublisher
	decision     DecisionEngine
	workerPool   WorkerPool
	logger       *slog.Logger
	metrics      *observability.Metrics
	publishAsync bool
}

func NewCreditService(
	creditRepo ports.CreditRepository,
	clientRepo ports.ClientRepository,
	bankRepo ports.BankRepository,
	eventBus ports.EventPublisher,
	decision DecisionEngine,
	workers int,
	logger *slog.Logger,
	metrics *observability.Metrics,
) *CreditService {
	return &CreditService{
		creditRepo:   creditRepo,
		clientRepo:   clientRepo,
		bankRepo:     bankRepo,
		eventBus:     eventBus,
		decision:     decision,
		workerPool:   NewWorkerPool(workers),
		logger:       logger,
		metrics:      metrics,
		publishAsync: true,
	}
}

func (s *CreditService) Create(ctx context.Context, input domain.Credit) (domain.Credit, error) {
	startedAt := time.Now()

	if input.ClientID == "" || input.BankID == "" {
		return domain.Credit{}, fmt.Errorf("%w: client_id and bank_id are required", domain.ErrValidation)
	}
	if err := input.CreditType.Validate(); err != nil {
		return domain.Credit{}, fmt.Errorf("%w: %v", domain.ErrValidation, err)
	}

	credit := domain.Credit{
		ID:         uuid.NewString(),
		ClientID:   input.ClientID,
		BankID:     input.BankID,
		MinPayment: input.MinPayment,
		MaxPayment: input.MaxPayment,
		TermMonths: input.TermMonths,
		CreditType: input.CreditType,
		CreatedAt:  time.Now().UTC(),
		Status:     domain.CreditStatusPending,
	}

	client, bank, status, err := s.resolveCreditDecision(ctx, credit)
	if err != nil {
		return domain.Credit{}, err
	}
	credit.Status = status

	if err := s.creditRepo.Create(ctx, credit); err != nil {
		return domain.Credit{}, err
	}

	s.publishCreditEvents(ctx, credit, client, bank)
	s.metrics.ObserveCreditDecision(time.Since(startedAt), string(credit.Status), string(credit.CreditType))
	s.logger.InfoContext(ctx,
		"credit created",
		slog.String("credit_id", credit.ID),
		slog.String("client_id", credit.ClientID),
		slog.String("bank_id", credit.BankID),
		slog.String("status", string(credit.Status)),
	)

	return credit, nil
}

func (s *CreditService) GetByID(ctx context.Context, id string) (domain.Credit, error) {
	return s.creditRepo.GetByID(ctx, id)
}

func (s *CreditService) List(ctx context.Context) ([]domain.Credit, error) {
	return s.creditRepo.List(ctx)
}

func (s *CreditService) resolveCreditDecision(ctx context.Context, credit domain.Credit) (domain.Client, domain.Bank, domain.CreditStatus, error) {
	validationTasks := []Task{
		{
			Name: "client",
			Run: func(ctx context.Context) (interface{}, error) {
				return s.clientRepo.GetByID(ctx, credit.ClientID)
			},
		},
		{
			Name: "bank",
			Run: func(ctx context.Context) (interface{}, error) {
				return s.bankRepo.GetByID(ctx, credit.BankID)
			},
		},
		{
			Name: "request_validation",
			Run: func(context.Context) (interface{}, error) {
				return nil, validateCreditConstraints(credit)
			},
		},
	}

	validationResults := s.workerPool.Run(ctx, validationTasks)
	for _, key := range []string{"client", "bank", "request_validation"} {
		result, ok := validationResults[key]
		if !ok {
			return domain.Client{}, domain.Bank{}, "", fmt.Errorf("%w: missing result from validation stage", domain.ErrUnavailable)
		}
		if result.Err != nil {
			return domain.Client{}, domain.Bank{}, "", result.Err
		}
	}

	client, ok := validationResults["client"].Value.(domain.Client)
	if !ok {
		return domain.Client{}, domain.Bank{}, "", errors.New("invalid client result type")
	}
	bank, ok := validationResults["bank"].Value.(domain.Bank)
	if !ok {
		return domain.Client{}, domain.Bank{}, "", errors.New("invalid bank result type")
	}

	scoreTasks := []Task{
		{
			Name: "term_score",
			Run: func(context.Context) (interface{}, error) {
				return scoreTerm(credit.TermMonths, credit.CreditType), nil
			},
		},
		{
			Name: "payment_score",
			Run: func(context.Context) (interface{}, error) {
				return scorePaymentRange(credit.MaxPayment - credit.MinPayment), nil
			},
		},
		{
			Name: "counterparty_score",
			Run: func(context.Context) (interface{}, error) {
				return scoreCounterparty(bank, client), nil
			},
		},
	}

	scoreResults := s.workerPool.Run(ctx, scoreTasks)
	for _, key := range []string{"term_score", "payment_score", "counterparty_score"} {
		result, ok := scoreResults[key]
		if !ok {
			return domain.Client{}, domain.Bank{}, "", fmt.Errorf("%w: missing result from scoring stage", domain.ErrUnavailable)
		}
		if result.Err != nil {
			return domain.Client{}, domain.Bank{}, "", result.Err
		}
	}

	termScore, ok := scoreResults["term_score"].Value.(float64)
	if !ok {
		return domain.Client{}, domain.Bank{}, "", errors.New("invalid term score result type")
	}
	paymentScore, ok := scoreResults["payment_score"].Value.(float64)
	if !ok {
		return domain.Client{}, domain.Bank{}, "", errors.New("invalid payment score result type")
	}
	counterpartyScore, ok := scoreResults["counterparty_score"].Value.(float64)
	if !ok {
		return domain.Client{}, domain.Bank{}, "", errors.New("invalid counterparty score result type")
	}

	totalScore := termScore + paymentScore + counterpartyScore
	status, err := s.decision.Evaluate(client, bank, credit, totalScore)
	if err != nil {
		return domain.Client{}, domain.Bank{}, "", err
	}

	return client, bank, status, nil
}

func (s *CreditService) publishCreditEvents(_ context.Context, credit domain.Credit, client domain.Client, bank domain.Bank) {
	events := []domain.Event{
		{
			Type:        "CreditCreated",
			AggregateID: credit.ID,
			OccurredAt:  time.Now().UTC(),
			Payload: map[string]interface{}{
				"credit": credit,
				"client": client,
				"bank":   bank,
			},
		},
	}
	if credit.Status == domain.CreditStatusApproved {
		events = append(events, domain.Event{
			Type:        "CreditApproved",
			AggregateID: credit.ID,
			OccurredAt:  time.Now().UTC(),
			Payload:     credit,
		})
	}

	publish := func() {
		publishCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		for _, event := range events {
			if err := s.eventBus.Publish(publishCtx, event); err != nil {
				s.logger.Warn("failed to publish event", slog.String("event_type", event.Type), slog.Any("error", err))
			}
		}
	}

	if s.publishAsync {
		go publish()
		return
	}

	publish()
}
