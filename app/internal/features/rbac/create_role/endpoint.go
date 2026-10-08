package createrole

import (
	"context"
	"errors"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/luiszkm/template-go/internal/features/rbac/create_role/db"
	"github.com/luiszkm/template-go/internal/features/rbac/role"
	"github.com/luiszkm/template-go/internal/platform/audit"
	platformdb "github.com/luiszkm/template-go/internal/platform/db"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
)

const permission op.Permission = "rbac:create"

const uniqueViolation = "23505"

type NewRole struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

type Input struct {
	Body NewRole
}

type Output struct {
	Body role.Role
}

var errNameTaken = errors.New("name taken")

func Register(api huma.API, d deps.Deps) error {
	return op.Register(api, op.Spec{
		ID:            "create-role",
		Method:        http.MethodPost,
		Path:          "/api/v1/rbac/roles",
		Summary:       "Create a role granting a subset of the permission catalogue",
		Tags:          []string{"rbac"},
		Errors:        []int{http.StatusConflict, http.StatusUnprocessableEntity},
		DefaultStatus: http.StatusCreated,
		Permission:    permission,
		AuditAction:   "role.created",
	}, func(ctx context.Context, in *Input) (*Output, error) {
		name, nameErr := role.NormalizeName(in.Body.Name)
		if err := role.Invalid(nameErr, role.CheckPermissions(api, in.Body.Permissions)); err != nil {
			return nil, err
		}
		permissions := slices.Sorted(slices.Values(in.Body.Permissions))
		if permissions == nil {
			permissions = []string{}
		}
		var created role.Role
		err := platformdb.WithTx(ctx, d.DB, func(tx pgx.Tx) error {
			q := db.New(tx)
			row, err := q.InsertRole(ctx, name)
			if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == uniqueViolation {
				return errNameTaken
			}
			if err != nil {
				return err
			}
			if err := q.InsertPermissions(ctx, db.InsertPermissionsParams{RoleID: row.ID, Permissions: permissions}); err != nil {
				return err
			}
			created = role.Role{ID: row.ID, Name: row.Name, Permissions: permissions, UserCount: 0}
			return audit.Record(ctx, tx, audit.Event{Action: "role.created", ResourceType: "role", ResourceID: row.ID.String(),
				After: role.Snapshot{Name: created.Name, Permissions: permissions}})
		})
		if errors.Is(err, errNameTaken) {
			return nil, huma.Error409Conflict("a role with this name already exists")
		}
		if err != nil {
			return nil, err
		}
		return &Output{Body: created}, nil
	})
}
