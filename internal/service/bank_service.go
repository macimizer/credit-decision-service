package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/macimizer/credit-decision-service/internal/domain"
	"github.com/macimizer/credit-decision-service/internal/ports"
)

type BankService struct {
	repo     ports.BankRepository
	cache    ports.Cache
	cacheTTL time.Duration
	logger   *slog.Logger
}

func NewBankService(repo ports.BankRepository, cache ports.Cache, cacheTTL time.Duration, logger *slog.Logger) *BankService {
	return &BankService{repo: repo, cache: cache, cacheTTL: cacheTTL, logger: logger}
}

func (s *BankService) Create(ctx context.Context, input domain.Bank) (domain.Bank, error) {
	if input.Name == "" {
		return domain.Bank{}, fmt.Errorf("%w: name is required", domain.ErrValidation)
	}
	if err := input.Type.Validate(); err != nil {
		return domain.Bank{}, fmt.Errorf("%w: %v", domain.ErrValidation, err)
	}

	bank := domain.Bank{
		ID:   uuid.NewString(),
		Name: input.Name,
		Type: input.Type,
	}

	if err := s.repo.Create(ctx, bank); err != nil {
		return domain.Bank{}, err
	}
	_ = s.cache.Delete(ctx, bankListCacheKey)
	_ = s.cache.Set(ctx, bankCacheKey(bank.ID), bank, s.cacheTTL)

	s.logger.InfoContext(ctx, "bank created", slog.String("bank_id", bank.ID))
	return bank, nil
}

func (s *BankService) GetByID(ctx context.Context, id string) (domain.Bank, error) {
	var bank domain.Bank
	if ok, err := s.cache.Get(ctx, bankCacheKey(id), &bank); err == nil && ok {
		return bank, nil
	}

	bank, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Bank{}, err
	}

	_ = s.cache.Set(ctx, bankCacheKey(id), bank, s.cacheTTL)
	return bank, nil
}

func (s *BankService) List(ctx context.Context) ([]domain.Bank, error) {
	var banks []domain.Bank
	if ok, err := s.cache.Get(ctx, bankListCacheKey, &banks); err == nil && ok {
		return banks, nil
	}

	banks, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, bankListCacheKey, banks, s.cacheTTL)
	return banks, nil
}

const bankListCacheKey = "banks:list"

func bankCacheKey(id string) string {
	return "banks:" + id
}
