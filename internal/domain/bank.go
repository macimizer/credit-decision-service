package domain

import "fmt"

type BankType string

const (
	BankTypePrivate    BankType = "PRIVATE"
	BankTypeGovernment BankType = "GOVERNMENT"
)

type Bank struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	Type BankType `json:"type"`
}

func (t BankType) Validate() error {
	switch t {
	case BankTypePrivate, BankTypeGovernment:
		return nil
	default:
		return fmt.Errorf("invalid bank type: %s", t)
	}
}
