package service

import (
	"fmt"
	"math"
	"time"

	"github.com/macimizer/credit-decision-service/internal/domain"
)

type DecisionEngine interface {
	Evaluate(client domain.Client, bank domain.Bank, credit domain.Credit, score float64) (domain.CreditStatus, error)
}

type RuleBasedDecisionEngine struct{}

func NewRuleBasedDecisionEngine() *RuleBasedDecisionEngine {
	return &RuleBasedDecisionEngine{}
}

func (e *RuleBasedDecisionEngine) Evaluate(client domain.Client, _ domain.Bank, credit domain.Credit, score float64) (domain.CreditStatus, error) {
	if err := validateCreditConstraints(credit); err != nil {
		return "", err
	}

	if client.Age(time.Now().UTC()) < 18 {
		return domain.CreditStatusRejected, nil
	}

	if math.Round(score) >= 60 {
		return domain.CreditStatusApproved, nil
	}

	return domain.CreditStatusRejected, nil
}

func validateCreditConstraints(credit domain.Credit) error {
	if credit.MinPayment <= 0 {
		return fmt.Errorf("%w: min_payment must be greater than zero", domain.ErrValidation)
	}
	if credit.MaxPayment < credit.MinPayment {
		return fmt.Errorf("%w: max_payment must be greater than or equal to min_payment", domain.ErrValidation)
	}
	if credit.TermMonths <= 0 {
		return fmt.Errorf("%w: term_months must be greater than zero", domain.ErrValidation)
	}

	switch credit.CreditType {
	case domain.CreditTypeAuto:
		if credit.TermMonths > 84 {
			return fmt.Errorf("%w: AUTO credits cannot exceed 84 months", domain.ErrValidation)
		}
	case domain.CreditTypeMortgage:
		if credit.TermMonths > 360 {
			return fmt.Errorf("%w: MORTGAGE credits cannot exceed 360 months", domain.ErrValidation)
		}
	case domain.CreditTypeCommercial:
		if credit.TermMonths > 180 {
			return fmt.Errorf("%w: COMMERCIAL credits cannot exceed 180 months", domain.ErrValidation)
		}
	default:
		return fmt.Errorf("%w: unsupported credit_type", domain.ErrValidation)
	}

	return nil
}

func scoreTerm(termMonths int, creditType domain.CreditType) float64 {
	switch creditType {
	case domain.CreditTypeMortgage:
		if termMonths <= 180 {
			return 35
		}
		if termMonths <= 300 {
			return 25
		}
		return 15
	default:
		if termMonths <= 12 {
			return 35
		}
		if termMonths <= 60 {
			return 28
		}
		return 18
	}
}

func scorePaymentRange(paymentSpread float64) float64 {
	switch {
	case paymentSpread <= 500:
		return 30
	case paymentSpread <= 2000:
		return 22
	case paymentSpread <= 5000:
		return 15
	default:
		return 8
	}
}

func scoreCounterparty(bank domain.Bank, client domain.Client) float64 {
	score := 10.0
	if bank.Type == domain.BankTypePrivate {
		score += 12
	} else {
		score += 8
	}

	age := client.Age(time.Now().UTC())
	if age >= 25 && age <= 65 {
		score += 18
	} else if age > 18 {
		score += 10
	}

	return score
}
