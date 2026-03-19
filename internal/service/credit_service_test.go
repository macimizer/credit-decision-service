package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/macimizer/credit-decision-service/internal/domain"
	"github.com/macimizer/credit-decision-service/internal/observability"
)

type mockClientRepo struct {
	client domain.Client
	err    error
}

func (m mockClientRepo) Create(context.Context, domain.Client) error { return nil }
func (m mockClientRepo) GetByID(context.Context, string) (domain.Client, error) {
	if m.err != nil {
		return domain.Client{}, m.err
	}
	return m.client, nil
}
func (m mockClientRepo) List(context.Context) ([]domain.Client, error) { return nil, nil }

type mockBankRepo struct {
	bank domain.Bank
	err  error
}

func (m mockBankRepo) Create(context.Context, domain.Bank) error { return nil }
func (m mockBankRepo) GetByID(context.Context, string) (domain.Bank, error) {
	if m.err != nil {
		return domain.Bank{}, m.err
	}
	return m.bank, nil
}
func (m mockBankRepo) List(context.Context) ([]domain.Bank, error) { return nil, nil }

type mockCreditRepo struct {
	created domain.Credit
	err     error
}

func (m *mockCreditRepo) Create(_ context.Context, credit domain.Credit) error {
	if m.err != nil {
		return m.err
	}
	m.created = credit
	return nil
}
func (m *mockCreditRepo) GetByID(context.Context, string) (domain.Credit, error) {
	return domain.Credit{}, nil
}
func (m *mockCreditRepo) List(context.Context) ([]domain.Credit, error) { return nil, nil }

type mockEventBus struct {
	events []domain.Event
	err    error
}

func (m *mockEventBus) Publish(_ context.Context, event domain.Event) error {
	if m.err != nil {
		return m.err
	}
	m.events = append(m.events, event)
	return nil
}

func TestCreditService_CreateApprovedCredit(t *testing.T) {
	creditRepo := &mockCreditRepo{}
	eventBus := &mockEventBus{}
	service := NewCreditService(
		creditRepo,
		mockClientRepo{client: domain.Client{ID: "client-1", BirthDate: time.Now().AddDate(-35, 0, 0)}},
		mockBankRepo{bank: domain.Bank{ID: "bank-1", Type: domain.BankTypePrivate}},
		eventBus,
		NewRuleBasedDecisionEngine(),
		4,
		slog.Default(),
		observability.NewMetrics(),
	)
	service.publishAsync = false

	credit, err := service.Create(context.Background(), domain.Credit{
		ClientID:   "client-1",
		BankID:     "bank-1",
		MinPayment: 100,
		MaxPayment: 450,
		TermMonths: 24,
		CreditType: domain.CreditTypeAuto,
	})

	require.NoError(t, err)
	require.Equal(t, domain.CreditStatusApproved, credit.Status)
	require.Equal(t, domain.CreditStatusApproved, creditRepo.created.Status)
	require.Len(t, eventBus.events, 2)
}

func TestCreditService_CreateReturnsErrorWhenClientMissing(t *testing.T) {
	creditRepo := &mockCreditRepo{}
	service := NewCreditService(
		creditRepo,
		mockClientRepo{err: domain.ErrNotFound},
		mockBankRepo{bank: domain.Bank{ID: "bank-1", Type: domain.BankTypePrivate}},
		&mockEventBus{},
		NewRuleBasedDecisionEngine(),
		4,
		slog.Default(),
		observability.NewMetrics(),
	)

	_, err := service.Create(context.Background(), domain.Credit{
		ClientID:   "client-1",
		BankID:     "bank-1",
		MinPayment: 100,
		MaxPayment: 450,
		TermMonths: 24,
		CreditType: domain.CreditTypeAuto,
	})

	require.Error(t, err)
	require.True(t, errors.Is(err, domain.ErrNotFound))
}
