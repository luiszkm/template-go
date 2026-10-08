package updaterole

import (
	"context"
	"errors"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/luiszkm/template-go/internal/features/rbac/role"
	"github.com/luiszkm/template-go/internal/features/rbac/update_role/db"
	"github.com/luiszkm/template-go/internal/platform/audit"
	platformdb "github.com/luiszkm/template-go/internal/platform/db"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
)

const permission op.Permission = "rbac:update"

const uniqueViolation = "23505"

type RoleChanges struct {
	Name        *string   `json:"name,omitempty"`
	Permissions *[]string `json:"permissions,omitempty"`
}

func (b *RoleChanges) Resolve(_ huma.Context, prefix *huma.PathBuffer) []error {
	if b.Name == nil && b.Permissions == nil {
		return []error{&huma.ErrorDetail{Location: prefix.String(), Message: "send name, permissions or both"}}
	}
	return nil
}

type Input struct {
	ID   uuid.UUID `path:"id"`
	Body RoleChanges
}

type Output struct {
	Body role.Role
}

var (
	errNotFound  = errors.New("not found")
	errAdmin     = errors.New("admin role")
	errNameTaken = errors.New("name taken")
)

func Register(api huma.API, d deps.Deps) error {
	return op.Register(api, op.Spec{
		ID:          "update-role",
		Method:      http.MethodPatch,
		Path:        "/api/v1/rbac/roles/{id}",
		Summary:     "Rename a role or replace the permissions it grants",
		Tags:        []string{"rbac"},
		Errors:      []int{http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity},
		Permission:  permission,
		AuditAction: "role.updated",
	}, func(ctx context.Context, in *Input) (*Output, error) {
		changes, err := validated(api, in.Body)
		if err != nil {
			return nil, err
		}
		var updated role.Role
		err = platformdb.WithTx(ctx, d.DB, func(tx pgx.Tx) error {
			var err error
			updated, err = apply(ctx, tx, in.ID, changes)
			return err
		})
		switch {
		case errors.Is(err, errNotFound):
			return nil, huma.Error404NotFound("role not found")
		case errors.Is(err, errAdmin):
			return nil, huma.Error409Conflict("the admin role cannot be changed")
		case errors.Is(err, errNameTaken):
			return nil, huma.Error409Conflict("a role with this name already exists")
		case err != nil:
			return nil, err
		}
		return &Output{Body: updated}, nil
	})
}

func validated(api huma.API, body RoleChanges) (RoleChanges, error) {
	var nameErr, permErr *huma.ErrorDetail
	if body.Name != nil {
		name, detail := role.NormalizeName(*body.Name)
		body.Name, nameErr = &name, detail
	}
	if body.Permissions != nil {
		permErr = role.CheckPermissions(api, *body.Permissions)
		sorted := nonNil(slices.Sorted(slices.Values(*body.Permissions)))
		body.Permissions = &sorted
	}
	return body, role.Invalid(nameErr, permErr)
}

func apply(ctx context.Context, tx pgx.Tx, id uuid.UUID, changes RoleChanges) (role.Role, error) {
	q := db.New(tx)
	current, err := q.LockRole(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return role.Role{}, errNotFound
	}
	if err != nil {
		return role.Role{}, err
	}
	if current.Name == role.Admin {
		return role.Role{}, errAdmin
	}
	before, err := q.RolePermissions(ctx, id)
	if err != nil {
		return role.Role{}, err
	}
	after := role.Role{ID: id, Name: current.Name, Permissions: nonNil(before)}
	if changes.Name != nil {
		err := q.RenameRole(ctx, db.RenameRoleParams{ID: id, Name: *changes.Name})
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == uniqueViolation {
			return role.Role{}, errNameTaken
		}
		if err != nil {
			return role.Role{}, err
		}
		after.Name = *changes.Name
	}
	if changes.Permissions != nil {
		if err := q.DeletePermissions(ctx, id); err != nil {
			return role.Role{}, err
		}
		if err := q.InsertPermissions(ctx, db.InsertPermissionsParams{RoleID: id, Permissions: *changes.Permissions}); err != nil {
			return role.Role{}, err
		}
		after.Permissions = *changes.Permissions
	}
	holders, err := q.CountHolders(ctx, id)
	if err != nil {
		return role.Role{}, err
	}
	after.UserCount = int(holders)
	return after, audit.Record(ctx, tx, audit.Event{Action: "role.updated", ResourceType: "role", ResourceID: id.String(),
		Before: role.Snapshot{Name: current.Name, Permissions: nonNil(before)},
		After:  role.Snapshot{Name: after.Name, Permissions: after.Permissions}})
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
