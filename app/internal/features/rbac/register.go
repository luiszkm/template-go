package rbac

import (
	"errors"

	"github.com/danielgtaylor/huma/v2"

	assignroles "github.com/luiszkm/template-go/internal/features/rbac/assign_roles"
	createrole "github.com/luiszkm/template-go/internal/features/rbac/create_role"
	deleterole "github.com/luiszkm/template-go/internal/features/rbac/delete_role"
	getrole "github.com/luiszkm/template-go/internal/features/rbac/get_role"
	getuserroles "github.com/luiszkm/template-go/internal/features/rbac/get_user_roles"
	listpermissions "github.com/luiszkm/template-go/internal/features/rbac/list_permissions"
	listroles "github.com/luiszkm/template-go/internal/features/rbac/list_roles"
	updaterole "github.com/luiszkm/template-go/internal/features/rbac/update_role"
	"github.com/luiszkm/template-go/internal/platform/deps"
	// slices:imports
)

func Register(api huma.API, d deps.Deps) error {
	return errors.Join(
		listpermissions.Register(api, d),
		listroles.Register(api, d),
		getrole.Register(api, d),
		createrole.Register(api, d),
		updaterole.Register(api, d),
		deleterole.Register(api, d),
		getuserroles.Register(api, d),
		assignroles.Register(api, d),
	// slices:register
	)
}
