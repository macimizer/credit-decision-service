package postgresrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/macimizer/credit-decision-service/internal/domain"
)

type BankRepository struct {
	db *sql.DB
}

func NewBankRepository(db *sql.DB) *BankRepository {
	return &BankRepository{db: db}
}

func (r *BankRepository) Create(ctx context.Context, bank domain.Bank) error {
	const query = `
		INSERT INTO banks (id, name, type)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.ExecContext(ctx, query, bank.ID, bank.Name, bank.Type)
	if err != nil {
		return fmt.Errorf("insert bank: %w", err)
	}

	return nil
}

func (r *BankRepository) GetByID(ctx context.Context, id string) (domain.Bank, error) {
	const query = `
		SELECT id, name, type
		FROM banks
		WHERE id = $1
	`

	var bank domain.Bank
	if err := r.db.QueryRowContext(ctx, query, id).Scan(&bank.ID, &bank.Name, &bank.Type); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Bank{}, domain.ErrNotFound
		}
		return domain.Bank{}, fmt.Errorf("get bank by id: %w", err)
	}

	return bank, nil
}

func (r *BankRepository) List(ctx context.Context) (banks []domain.Bank, err error) {
	const query = `
		SELECT id, name, type
		FROM banks
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list banks: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			if err == nil {
				err = fmt.Errorf("close banks rows: %w", closeErr)
			} else {
				err = fmt.Errorf("%w; close banks rows: %v", err, closeErr)
			}
		}
	}()

	banks = make([]domain.Bank, 0)
	for rows.Next() {
		var bank domain.Bank
		if scanErr := rows.Scan(&bank.ID, &bank.Name, &bank.Type); scanErr != nil {
			return nil, fmt.Errorf("scan bank: %w", scanErr)
		}
		banks = append(banks, bank)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate banks: %w", err)
	}

	return banks, nil
}
