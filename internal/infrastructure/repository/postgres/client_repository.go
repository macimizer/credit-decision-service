package postgresrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/macimizer/credit-decision-service/internal/domain"
)

type ClientRepository struct {
	db *sql.DB
}

func NewClientRepository(db *sql.DB) *ClientRepository {
	return &ClientRepository{db: db}
}

func (r *ClientRepository) Create(ctx context.Context, client domain.Client) error {
	const query = `
		INSERT INTO clients (id, full_name, email, birth_date, country, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		client.ID,
		client.FullName,
		client.Email,
		client.BirthDate,
		client.Country,
		client.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert client: %w", err)
	}

	return nil
}

func (r *ClientRepository) GetByID(ctx context.Context, id string) (domain.Client, error) {
	const query = `
		SELECT id, full_name, email, birth_date, country, created_at
		FROM clients
		WHERE id = $1
	`

	var client domain.Client
	if err := r.db.QueryRowContext(ctx, query, id).Scan(
		&client.ID,
		&client.FullName,
		&client.Email,
		&client.BirthDate,
		&client.Country,
		&client.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Client{}, domain.ErrNotFound
		}
		return domain.Client{}, fmt.Errorf("get client by id: %w", err)
	}

	return client, nil
}

func (r *ClientRepository) List(ctx context.Context) (clients []domain.Client, err error) {
	const query = `
		SELECT id, full_name, email, birth_date, country, created_at
		FROM clients
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list clients: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			if err == nil {
				err = fmt.Errorf("close clients rows: %w", closeErr)
			} else {
				err = fmt.Errorf("%w; close clients rows: %v", err, closeErr)
			}
		}
	}()

	clients = make([]domain.Client, 0)
	for rows.Next() {
		var client domain.Client
		if scanErr := rows.Scan(
			&client.ID,
			&client.FullName,
			&client.Email,
			&client.BirthDate,
			&client.Country,
			&client.CreatedAt,
		); scanErr != nil {
			return nil, fmt.Errorf("scan client: %w", scanErr)
		}
		clients = append(clients, client)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate clients: %w", err)
	}

	return clients, nil
}
