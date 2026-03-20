package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/macimizer/credit-decision-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

type RedisStreamPublisher struct {
	client     *redis.Client
	streamName string
}

func NewRedisStreamPublisher(client *redis.Client, streamName string) *RedisStreamPublisher {
	return &RedisStreamPublisher{client: client, streamName: streamName}
}

func (p *RedisStreamPublisher) Publish(ctx context.Context, event domain.Event) error {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("marshal event payload: %w", err)
	}

	return p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: p.streamName,
		Values: map[string]interface{}{
			"type":         event.Type,
			"aggregate_id": event.AggregateID,
			"occurred_at":  event.OccurredAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
			"payload":      payload,
		},
	}).Err()
}
