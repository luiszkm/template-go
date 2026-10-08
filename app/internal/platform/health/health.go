package health

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"

	"github.com/luiszkm/template-go/internal/platform/op"
)

type DB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Status struct {
	Status string `json:"status" enum:"ok,ready"`
}

type statusOutput struct{ Body Status }

func Register(api huma.API, db DB, timeout time.Duration) error {
	if err := op.Register(api, op.Spec{
		ID: "healthz", Method: http.MethodGet, Path: "/healthz", Tags: []string{"health"},
		Summary: "Liveness: the process is up", Public: true,
	}, func(context.Context, *struct{}) (*statusOutput, error) {
		return &statusOutput{Body: Status{Status: "ok"}}, nil
	}); err != nil {
		return err
	}
	return op.Register(api, op.Spec{
		ID: "readyz", Method: http.MethodGet, Path: "/readyz", Tags: []string{"health"},
		Summary: "Readiness: the database answers", Public: true,
	}, func(ctx context.Context, _ *struct{}) (*statusOutput, error) {
		if db == nil {
			return nil, huma.Error503ServiceUnavailable("database not configured")
		}
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		if err := ping(ctx, db); err != nil {
			return nil, huma.Error503ServiceUnavailable("database unavailable")
		}
		return &statusOutput{Body: Status{Status: "ready"}}, nil
	})
}

func ping(ctx context.Context, db DB) error {
	done := make(chan error, 1)
	go func() {
		var one int
		done <- db.QueryRow(ctx, "SELECT 1").Scan(&one)
	}()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
