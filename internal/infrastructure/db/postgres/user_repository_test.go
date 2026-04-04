package postgres

import (
	"context"
	"testing"
	_ "time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/tazhibayda/OrbittoAuth/internal/domain/user"
)

func TestUserRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://postgres:1234@192.168.32.123:5432/test_db?sslmode=disable")
	assert.NoError(t, err)
	defer pool.Close()

	repo := NewUserRepository(pool)

	u, _ := user.NewUser("test-integration@mail.com", "Password123!", "Ivan", "Ivanov")

	err = repo.Create(ctx, u)
	assert.NoError(t, err)

	found, err := repo.GetByEmail(ctx, u.Email)
	assert.NoError(t, err)
	assert.Equal(t, u.ID, found.ID)
	assert.Equal(t, u.FirstName, found.FirstName)

	pool.Exec(ctx, "DELETE FROM users WHERE id = $1", u.ID)
}
