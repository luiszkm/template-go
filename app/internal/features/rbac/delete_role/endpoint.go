package deleterole

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/luiszkm/template-go/internal/features/rbac/delete_role/db"
	"github.com/luiszkm/template-go/internal/features/rbac/role"
	"github.com/luiszkm/template-go/internal/platform/audit"
	platformdb "github.com/luiszkm/template-go/internal/platform/db"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
)

const permission op.Permission = "rbac:delete"

type Input struct {
	ID uuid.UUID `path:"id"`
}

var (
	errNotFound = errors.New("not found")
	errAdmin    = errors.New("admin role")
	errInUse    = errors.New("role in use")
)

func Register(api huma.API, d deps.Deps) error {
	return op.Register(api, op.Spec{
		ID:            "delete-role",
		Method:        http.MethodDelete,
		Path:          "/api/v1/rbac/roles/{id}",
		Summary:       "Delete a role that no user holds",
		Tags:          []string{"rbac"},
		Errors:        []int{http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity},
		DefaultStatus: http.StatusNoContent,
		Permission:    permission,
		AuditAction:   "role.deleted",
	}, func(ctx context.Context, in *Input) (*struct{}, error) {
		err := platformdb.WithTx(ctx, d.DB, func(tx pgx.Tx) error {
			return remove(ctx, tx, in.ID)
		})
		switch {
		case errors.Is(err, errNotFound):
			return nil, huma.Error404NotFound("role not found")
		case errors.Is(err, errAdmin):
			return nil, huma.Error409Conflict("the admin role cannot be deleted")
		case errors.Is(err, errInUse):
			return nil, huma.Error409Conflict("the role is held by at least one user")
		}
		return nil, err
	})
}

func remove(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	q := db.New(tx)
	current, err := q.LockRole(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return errNotFound
	}
	if err != nil {
		return err
	}
	if current.Name == role.Admin {
		return errAdmin
	}
	holders, err := q.CountHolders(ctx, id)
	if err != nil {
		return err
	}
	if holders > 0 {
		return errInUse
	}
	permissions, err := q.RolePermissions(ctx, id)
	if err != nil {
		return err
	}
	if permissions == nil {
		permissions = []string{}
	}
	if err := q.DeleteRole(ctx, id); err != nil {
		return err
	}
	return audit.Record(ctx, tx, audit.Event{Action: "role.deleted", ResourceType: "role", ResourceID: id.String(),
		Before: role.Snapshot{Name: current.Name, Permissions: permissions}})
}
