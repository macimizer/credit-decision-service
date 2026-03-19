package postgresrepo

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/macimizer/credit-decision-service/internal/domain"
)

func TestClientRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewClientRepository(db)
	birthDate := time.Date(1990, 5, 12, 0, 0, 0, 0, time.UTC)
	createdAt := time.Now().UTC()

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, full_name, email, birth_date, country, created_at
		FROM clients
		WHERE id = $1
	`)).WithArgs("client-1").WillReturnRows(sqlmock.NewRows([]string{"id", "full_name", "email", "birth_date", "country", "created_at"}).
		AddRow("client-1", "Jane Doe", "jane@example.com", birthDate, "CO", createdAt))

	client, err := repo.GetByID(context.Background(), "client-1")
	require.NoError(t, err)
	require.Equal(t, domain.Client{ID: "client-1", FullName: "Jane Doe", Email: "jane@example.com", BirthDate: birthDate, Country: "CO", CreatedAt: createdAt}, client)
	require.NoError(t, mock.ExpectationsWereMet())
}
