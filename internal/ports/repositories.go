package ports

import (
	"context"
	"time"

	"github.com/macimizer/credit-decision-service/internal/domain"
)

type ClientRepository interface {
	Create(ctx context.Context, client domain.Client) error
	GetByID(ctx context.Context, id string) (domain.Client, error)
	List(ctx context.Context) ([]domain.Client, error)
}

type BankRepository interface {
	Create(ctx context.Context, bank domain.Bank) error
	GetByID(ctx context.Context, id string) (domain.Bank, error)
	List(ctx context.Context) ([]domain.Bank, error)
}

type CreditRepository interface {
	Create(ctx context.Context, credit domain.Credit) error
	GetByID(ctx context.Context, id string) (domain.Credit, error)
	List(ctx context.Context) ([]domain.Credit, error)
}

type Cache interface {
	Get(ctx context.Context, key string, destination interface{}) (bool, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

type EventPublisher interface {
	Publish(ctx context.Context, event domain.Event) error
}
