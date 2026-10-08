package features

import (
	"errors"

	"github.com/danielgtaylor/huma/v2"

	"github.com/luiszkm/template-go/internal/features/users"
	"github.com/luiszkm/template-go/internal/platform/deps"
	// features:imports
)

func Register(api huma.API, d deps.Deps) error {
	return errors.Join(
		users.Register(api, d),
	// features:register
	)
}
