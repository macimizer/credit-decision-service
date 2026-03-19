//go:build integration

package integration

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"github.com/macimizer/credit-decision-service/internal/domain"
	postgresrepo "github.com/macimizer/credit-decision-service/internal/infrastructure/repository/postgres"
	postgresplatform "github.com/macimizer/credit-decision-service/internal/platform/postgres"
)

func TestClientAndCreditRepositories(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("set RUN_INTEGRATION_TESTS=true to execute integration tests")
	}

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Fatal("TEST_DATABASE_DSN must be set")
	}

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	require.NoError(t, postgresplatform.Migrate(ctx, db))
	cleanupTables(t, db)

	clientRepo := postgresrepo.NewClientRepository(db)
	bankRepo := postgresrepo.NewBankRepository(db)
	creditRepo := postgresrepo.NewCreditRepository(db)

	client := domain.Client{
		ID:        uuid.NewString(),
		FullName:  "Integration User",
		Email:     "integration@example.com",
		BirthDate: time.Date(1990, 8, 10, 0, 0, 0, 0, time.UTC),
		Country:   "BR",
		CreatedAt: time.Now().UTC(),
	}
	require.NoError(t, clientRepo.Create(ctx, client))

	bank := domain.Bank{
		ID:   uuid.NewString(),
		Name: "National Bank",
		Type: domain.BankTypeGovernment,
	}
	require.NoError(t, bankRepo.Create(ctx, bank))

	credit := domain.Credit{
		ID:         uuid.NewString(),
		ClientID:   client.ID,
		BankID:     bank.ID,
		MinPayment: 100,
		MaxPayment: 450,
		TermMonths: 12,
		CreditType: domain.CreditTypeAuto,
		CreatedAt:  time.Now().UTC(),
		Status:     domain.CreditStatusApproved,
	}
	require.NoError(t, creditRepo.Create(ctx, credit))

	storedClient, err := clientRepo.GetByID(ctx, client.ID)
	require.NoError(t, err)
	require.Equal(t, client.Email, storedClient.Email)

	storedCredit, err := creditRepo.GetByID(ctx, credit.ID)
	require.NoError(t, err)
	require.Equal(t, domain.CreditStatusApproved, storedCredit.Status)
}

func cleanupTables(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`TRUNCATE TABLE credits, banks, clients RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
}
