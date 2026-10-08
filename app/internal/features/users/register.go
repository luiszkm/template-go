package users

import (
	"errors"

	"github.com/danielgtaylor/huma/v2"

	activateuser "github.com/luiszkm/template-go/internal/features/users/activate_user"
	changepassword "github.com/luiszkm/template-go/internal/features/users/change_password"
	createuser "github.com/luiszkm/template-go/internal/features/users/create_user"
	deactivateuser "github.com/luiszkm/template-go/internal/features/users/deactivate_user"
	getuser "github.com/luiszkm/template-go/internal/features/users/get_user"
	listusers "github.com/luiszkm/template-go/internal/features/users/list_users"
	"github.com/luiszkm/template-go/internal/features/users/login"
	"github.com/luiszkm/template-go/internal/features/users/logout"
	"github.com/luiszkm/template-go/internal/features/users/me"
	updateuser "github.com/luiszkm/template-go/internal/features/users/update_user"
	"github.com/luiszkm/template-go/internal/platform/deps"
	// slices:imports
)

func Register(api huma.API, d deps.Deps) error {
	return errors.Join(
		login.Register(api, d),
		logout.Register(api, d),
		me.Register(api, d),
		changepassword.Register(api, d),
		createuser.Register(api, d),
		listusers.Register(api, d),
		getuser.Register(api, d),
		updateuser.Register(api, d),
		deactivateuser.Register(api, d),
		activateuser.Register(api, d),
	// slices:register
	)
}
