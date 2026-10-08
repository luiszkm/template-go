package audit

import (
	"errors"

	"github.com/danielgtaylor/huma/v2"

	getevent "github.com/luiszkm/template-go/internal/features/audit/get_event"
	listactions "github.com/luiszkm/template-go/internal/features/audit/list_actions"
	listevents "github.com/luiszkm/template-go/internal/features/audit/list_events"
	"github.com/luiszkm/template-go/internal/platform/deps"
	// slices:imports
)

func Register(api huma.API, d deps.Deps) error {
	return errors.Join(
		listevents.Register(api, d),
		getevent.Register(api, d),
		listactions.Register(api, d),
	// slices:register
	)
}
