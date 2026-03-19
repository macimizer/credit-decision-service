package service

import (
	"testing"
	"time"

	"github.com/macimizer/credit-decision-service/internal/domain"
)

func BenchmarkRuleBasedDecisionEngine_Evaluate(b *testing.B) {
	engine := NewRuleBasedDecisionEngine()
	client := domain.Client{BirthDate: time.Now().AddDate(-32, 0, 0)}
	bank := domain.Bank{Type: domain.BankTypePrivate}
	credit := domain.Credit{MinPayment: 100, MaxPayment: 600, TermMonths: 24, CreditType: domain.CreditTypeAuto}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Evaluate(client, bank, credit, 72)
	}
}
