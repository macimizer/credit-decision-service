package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/mail"
	"time"

	"github.com/google/uuid"
	"github.com/macimizer/credit-decision-service/internal/domain"
	"github.com/macimizer/credit-decision-service/internal/ports"
)

type ClientService struct {
	repo     ports.ClientRepository
	cache    ports.Cache
	cacheTTL time.Duration
	logger   *slog.Logger
}

func NewClientService(repo ports.ClientRepository, cache ports.Cache, cacheTTL time.Duration, logger *slog.Logger) *ClientService {
	return &ClientService{repo: repo, cache: cache, cacheTTL: cacheTTL, logger: logger}
}

func (s *ClientService) Create(ctx context.Context, input domain.Client) (domain.Client, error) {
	if input.FullName == "" {
		return domain.Client{}, fmt.Errorf("%w: full_name is required", domain.ErrValidation)
	}
	if _, err := mail.ParseAddress(input.Email); err != nil {
		return domain.Client{}, fmt.Errorf("%w: invalid email", domain.ErrValidation)
	}
	if input.Country == "" {
		return domain.Client{}, fmt.Errorf("%w: country is required", domain.ErrValidation)
	}
	if input.BirthDate.IsZero() {
		return domain.Client{}, fmt.Errorf("%w: birth_date is required", domain.ErrValidation)
	}

	client := domain.Client{
		ID:        uuid.NewString(),
		FullName:  input.FullName,
		Email:     input.Email,
		BirthDate: input.BirthDate.UTC(),
		Country:   input.Country,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, client); err != nil {
		return domain.Client{}, err
	}
	_ = s.cache.Delete(ctx, clientListCacheKey)
	_ = s.cache.Set(ctx, clientCacheKey(client.ID), client, s.cacheTTL)

	s.logger.InfoContext(ctx, "client created", slog.String("client_id", client.ID))
	return client, nil
}

func (s *ClientService) GetByID(ctx context.Context, id string) (domain.Client, error) {
	var client domain.Client
	if ok, err := s.cache.Get(ctx, clientCacheKey(id), &client); err == nil && ok {
		return client, nil
	}

	client, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Client{}, err
	}

	_ = s.cache.Set(ctx, clientCacheKey(id), client, s.cacheTTL)
	return client, nil
}

func (s *ClientService) List(ctx context.Context) ([]domain.Client, error) {
	var clients []domain.Client
	if ok, err := s.cache.Get(ctx, clientListCacheKey, &clients); err == nil && ok {
		return clients, nil
	}

	clients, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, clientListCacheKey, clients, s.cacheTTL)
	return clients, nil
}

const clientListCacheKey = "clients:list"

func clientCacheKey(id string) string {
	return "clients:" + id
}
