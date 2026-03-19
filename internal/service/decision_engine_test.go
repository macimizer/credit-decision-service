package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/macimizer/credit-decision-service/internal/domain"
)

func TestRuleBasedDecisionEngine_EvaluateApproved(t *testing.T) {
	engine := NewRuleBasedDecisionEngine()

	status, err := engine.Evaluate(
		domain.Client{BirthDate: time.Now().AddDate(-30, 0, 0)},
		domain.Bank{Type: domain.BankTypePrivate},
		domain.Credit{MinPayment: 100, MaxPayment: 500, TermMonths: 24, CreditType: domain.CreditTypeAuto},
		80,
	)

	require.NoError(t, err)
	require.Equal(t, domain.CreditStatusApproved, status)
}

func TestRuleBasedDecisionEngine_EvaluateRejectsUnderage(t *testing.T) {
	engine := NewRuleBasedDecisionEngine()

	status, err := engine.Evaluate(
		domain.Client{BirthDate: time.Now().AddDate(-16, 0, 0)},
		domain.Bank{Type: domain.BankTypePrivate},
		domain.Credit{MinPayment: 100, MaxPayment: 500, TermMonths: 24, CreditType: domain.CreditTypeAuto},
		80,
	)

	require.NoError(t, err)
	require.Equal(t, domain.CreditStatusRejected, status)
}

func TestRuleBasedDecisionEngine_ValidateConstraints(t *testing.T) {
	engine := NewRuleBasedDecisionEngine()

	_, err := engine.Evaluate(
		domain.Client{BirthDate: time.Now().AddDate(-40, 0, 0)},
		domain.Bank{Type: domain.BankTypePrivate},
		domain.Credit{MinPayment: 500, MaxPayment: 100, TermMonths: 24, CreditType: domain.CreditTypeAuto},
		80,
	)

	require.Error(t, err)
}
