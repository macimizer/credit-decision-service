package domain

import (
	"fmt"
	"time"
)

type CreditType string

type CreditStatus string

const (
	CreditTypeAuto       CreditType = "AUTO"
	CreditTypeMortgage   CreditType = "MORTGAGE"
	CreditTypeCommercial CreditType = "COMMERCIAL"
)

const (
	CreditStatusPending  CreditStatus = "PENDING"
	CreditStatusApproved CreditStatus = "APPROVED"
	CreditStatusRejected CreditStatus = "REJECTED"
)

type Credit struct {
	ID         string       `json:"id"`
	ClientID   string       `json:"client_id"`
	BankID     string       `json:"bank_id"`
	MinPayment float64      `json:"min_payment"`
	MaxPayment float64      `json:"max_payment"`
	TermMonths int          `json:"term_months"`
	CreditType CreditType   `json:"credit_type"`
	CreatedAt  time.Time    `json:"created_at"`
	Status     CreditStatus `json:"status"`
}

func (t CreditType) Validate() error {
	switch t {
	case CreditTypeAuto, CreditTypeMortgage, CreditTypeCommercial:
		return nil
	default:
		return fmt.Errorf("invalid credit type: %s", t)
	}
}

func (s CreditStatus) Validate() error {
	switch s {
	case CreditStatusPending, CreditStatusApproved, CreditStatusRejected:
		return nil
	default:
		return fmt.Errorf("invalid credit status: %s", s)
	}
}
