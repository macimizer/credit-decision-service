package domain

import "time"

type Event struct {
	Type        string      `json:"type"`
	AggregateID string      `json:"aggregate_id"`
	OccurredAt  time.Time   `json:"occurred_at"`
	Payload     interface{} `json:"payload"`
}
