package assignroles

import (
	"context"
	"errors"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/luiszkm/template-go/internal/features/rbac/assign_roles/db"
	"github.com/luiszkm/template-go/internal/features/rbac/role"
	"github.com/luiszkm/template-go/internal/platform/audit"
	platformdb "github.com/luiszkm/template-go/internal/platform/db"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
)

const permission op.Permission = "rbac:assign"

const foreignKeyViolation = "23503"

type RoleAssignment struct {
	RoleIDs []uuid.UUID `json:"role_ids"`
}

func (b *RoleAssignment) Resolve(_ huma.Context, prefix *huma.PathBuffer) []error {
	seen := map[uuid.UUID]bool{}
	for _, id := range b.RoleIDs {
		if seen[id] {
			return []error{&huma.ErrorDetail{Location: prefix.With("role_ids"), Message: "duplicate role id " + id.String(), Value: id}}
		}
		seen[id] = true
	}
	return nil
}

type Input struct {
	ID   uuid.UUID `path:"id"`
	Body RoleAssignment
}

type held struct {
	Roles []string `json:"roles"`
}

var (
	errNotFound    = errors.New("not found")
	errUnknownRole = errors.New("unknown role")
	errLastAdmin   = errors.New("last admin")
)

func Register(api huma.API, d deps.Deps) error {
	return op.Register(api, op.Spec{
		ID:            "assign-roles",
		Method:        http.MethodPut,
		Path:          "/api/v1/rbac/users/{id}/roles",
		Summary:       "Replace the roles of a user and end all of their sessions",
		Tags:          []string{"rbac"},
		Errors:        []int{http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity},
		DefaultStatus: http.StatusNoContent,
		Permission:    permission,
		AuditAction:   "user.roles_changed",
	}, func(ctx context.Context, in *Input) (*struct{}, error) {
		err := platformdb.WithTx(ctx, d.DB, func(tx pgx.Tx) error {
			return assign(ctx, tx, in.ID, in.Body.RoleIDs)
		})
		switch {
		case errors.Is(err, errNotFound):
			return nil, huma.Error404NotFound("user not found")
		case errors.Is(err, errUnknownRole):
			return nil, huma.Error422UnprocessableEntity("validation failed",
				&huma.ErrorDetail{Location: "body.role_ids", Message: "every role id must name an existing role"})
		case errors.Is(err, errLastAdmin):
			return nil, huma.Error409Conflict("no active user would hold the admin role")
		}
		return nil, err
	})
}

func assign(ctx context.Context, tx pgx.Tx, userID uuid.UUID, roleIDs []uuid.UUID) error {
	q := db.New(tx)
	if _, err := q.LockUser(ctx, userID); errors.Is(err, pgx.ErrNoRows) {
		return errNotFound
	} else if err != nil {
		return err
	}
	wanted, err := q.RolesByID(ctx, roleIDs)
	if err != nil {
		return err
	}
	if len(wanted) != len(roleIDs) {
		return errUnknownRole
	}
	current, err := q.UserRoles(ctx, userID)
	if err != nil {
		return err
	}
	before, after := names(current), names(wanted)
	if slices.Equal(before, after) {
		return nil
	}
	if slices.Contains(before, role.Admin) && !slices.Contains(after, role.Admin) {
		if err := keepAnActiveAdmin(ctx, q, userID); err != nil {
			return err
		}
	}
	if err := q.ClearUserRoles(ctx, userID); err != nil {
		return err
	}
	err = q.InsertUserRoles(ctx, db.InsertUserRolesParams{UserID: userID, RoleIds: roleIDs})
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == foreignKeyViolation {
		return errUnknownRole
	}
	if err != nil {
		return err
	}
	if _, err := q.DeleteUserSessions(ctx, userID); err != nil {
		return err
	}
	return audit.Record(ctx, tx, audit.Event{Action: "user.roles_changed", ResourceType: "user", ResourceID: userID.String(),
		Before: held{Roles: before}, After: held{Roles: after}})
}

func keepAnActiveAdmin(ctx context.Context, q *db.Queries, userID uuid.UUID) error {
	adminID, err := q.LockAdminRole(ctx)
	if err != nil {
		return err
	}
	others, err := q.CountOtherActiveHolders(ctx, db.CountOtherActiveHoldersParams{RoleID: adminID, UserID: userID})
	if err != nil {
		return err
	}
	if others == 0 {
		return errLastAdmin
	}
	return nil
}

func names(roles []db.Role) []string {
	out := make([]string, 0, len(roles))
	for _, r := range roles {
		out = append(out, r.Name)
	}
	slices.Sort(out)
	return out
}
