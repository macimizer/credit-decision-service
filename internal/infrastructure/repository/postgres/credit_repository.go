package postgresrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/macimizer/credit-decision-service/internal/domain"
)

type CreditRepository struct {
	db *sql.DB
}

func NewCreditRepository(db *sql.DB) *CreditRepository {
	return &CreditRepository{db: db}
}

func (r *CreditRepository) Create(ctx context.Context, credit domain.Credit) error {
	const query = `
		INSERT INTO credits (id, client_id, bank_id, min_payment, max_payment, term_months, credit_type, created_at, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		credit.ID,
		credit.ClientID,
		credit.BankID,
		credit.MinPayment,
		credit.MaxPayment,
		credit.TermMonths,
		credit.CreditType,
		credit.CreatedAt,
		credit.Status,
	)
	if err != nil {
		return fmt.Errorf("insert credit: %w", err)
	}

	return nil
}

func (r *CreditRepository) GetByID(ctx context.Context, id string) (domain.Credit, error) {
	const query = `
		SELECT id, client_id, bank_id, min_payment, max_payment, term_months, credit_type, created_at, status
		FROM credits
		WHERE id = $1
	`

	var credit domain.Credit
	if err := r.db.QueryRowContext(ctx, query, id).Scan(
		&credit.ID,
		&credit.ClientID,
		&credit.BankID,
		&credit.MinPayment,
		&credit.MaxPayment,
		&credit.TermMonths,
		&credit.CreditType,
		&credit.CreatedAt,
		&credit.Status,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Credit{}, domain.ErrNotFound
		}
		return domain.Credit{}, fmt.Errorf("get credit by id: %w", err)
	}

	return credit, nil
}

func (r *CreditRepository) List(ctx context.Context) (credits []domain.Credit, err error) {
	const query = `
		SELECT id, client_id, bank_id, min_payment, max_payment, term_months, credit_type, created_at, status
		FROM credits
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list credits: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			if err == nil {
				err = fmt.Errorf("close credits rows: %w", closeErr)
			} else {
				err = fmt.Errorf("%w; close credits rows: %v", err, closeErr)
			}
		}
	}()

	credits = make([]domain.Credit, 0)
	for rows.Next() {
		var credit domain.Credit
		if scanErr := rows.Scan(
			&credit.ID,
			&credit.ClientID,
			&credit.BankID,
			&credit.MinPayment,
			&credit.MaxPayment,
			&credit.TermMonths,
			&credit.CreditType,
			&credit.CreatedAt,
			&credit.Status,
		); scanErr != nil {
			return nil, fmt.Errorf("scan credit: %w", scanErr)
		}
		credits = append(credits, credit)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate credits: %w", err)
	}

	return credits, nil
}
