package getevent

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"

	"github.com/luiszkm/template-go/internal/features/audit/event"
	"github.com/luiszkm/template-go/internal/features/audit/get_event/db"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
)

const permission op.Permission = "audit:read"

type Input struct {
	ID int64 `path:"id" minimum:"1"`
}

type AuditEventDetail struct {
	event.AuditEvent
	Before json.RawMessage `json:"before" nullable:"true"`
	After  json.RawMessage `json:"after" nullable:"true"`
}

type Output struct {
	Body AuditEventDetail
}

func Register(api huma.API, d deps.Deps) error {
	return op.Register(api, op.Spec{
		ID:         "get-audit-event",
		Method:     http.MethodGet,
		Path:       "/api/v1/audit/events/{id}",
		Summary:    "One audit event with the state before and after the change",
		Tags:       []string{"audit"},
		Errors:     []int{http.StatusNotFound, http.StatusUnprocessableEntity},
		Permission: permission,
	}, func(ctx context.Context, in *Input) (*Output, error) {
		r, err := db.New(d.DB).GetEvent(ctx, in.ID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("audit event not found")
		}
		if err != nil {
			return nil, err
		}
		return &Output{Body: AuditEventDetail{
			AuditEvent: event.From(event.Row{ID: r.ID, OccurredAt: r.OccurredAt, Action: r.Action, ActorID: r.ActorID,
				ActorEmail: r.ActorEmail, ResourceType: r.ResourceType, ResourceID: r.ResourceID, IP: r.Ip, RequestID: r.RequestID}),
			Before: orNull(r.Before),
			After:  orNull(r.After),
		}}, nil
	})
}

func orNull(raw []byte) json.RawMessage {
	if raw == nil {
		return json.RawMessage("null")
	}
	return raw
}
