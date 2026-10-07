// Package features lists every feature. `task new:slice` edits the marked lines; keep them.
package features

import (
	"errors"

	"github.com/danielgtaylor/huma/v2"

	"github.com/luiszkm/template-go/internal/platform/deps"
	// features:imports
)

// Register registers every feature's operations.
func Register(api huma.API, d deps.Deps) error {
	return errors.Join(
	// features:register
	)
}
