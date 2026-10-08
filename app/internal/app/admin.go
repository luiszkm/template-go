package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/luiszkm/template-go/internal/features/users/bootstrap"
)

func ValidateAdmin(email, name, password string) error {
	return bootstrap.Validate(email, name, password)
}

func CreateAdmin(ctx context.Context, pool *pgxpool.Pool, email, name, password string) (uuid.UUID, error) {
	return bootstrap.CreateAdmin(ctx, pool, email, name, password)
}
