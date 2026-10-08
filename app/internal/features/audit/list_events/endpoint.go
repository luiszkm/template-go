package listevents

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/luiszkm/template-go/internal/features/audit/event"
	"github.com/luiszkm/template-go/internal/features/audit/list_events/db"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
)

const permission op.Permission = "audit:read"

type Input struct {
	Limit        int32  `query:"limit" default:"50" minimum:"1" maximum:"100"`
	Before       int64  `query:"before" minimum:"1"`
	Action       string `query:"action"`
	ActorID      string `query:"actor_id" format:"uuid"`
	ResourceType string `query:"resource_type"`
	ResourceID   string `query:"resource_id"`
	From         string `query:"from" format:"date-time"`
	To           string `query:"to" format:"date-time"`
}

type AuditEventPage struct {
	Items []event.AuditEvent `json:"items"`
	Next  *int64             `json:"next" nullable:"true"`
}

type Output struct {
	Body AuditEventPage
}

func Register(api huma.API, d deps.Deps) error {
	return op.Register(api, op.Spec{
		ID:         "list-audit-events",
		Method:     http.MethodGet,
		Path:       "/api/v1/audit/events",
		Summary:    "Audit events, newest first, filtered and paged by cursor",
		Tags:       []string{"audit"},
		Errors:     []int{http.StatusUnprocessableEntity},
		Permission: permission,
	}, func(ctx context.Context, in *Input) (*Output, error) {
		params, err := paramsOf(in)
		if err != nil {
			return nil, err
		}
		rows, err := db.New(d.DB).ListEvents(ctx, params)
		if err != nil {
			return nil, err
		}
		page := AuditEventPage{Items: make([]event.AuditEvent, 0, len(rows))}
		for i, r := range rows {
			if i == int(in.Limit) {
				last := page.Items[len(page.Items)-1].ID
				page.Next = &last
				break
			}
			page.Items = append(page.Items, event.From(event.Row{ID: r.ID, OccurredAt: r.OccurredAt, Action: r.Action, ActorID: r.ActorID,
				ActorEmail: r.ActorEmail, ResourceType: r.ResourceType, ResourceID: r.ResourceID, IP: r.Ip, RequestID: r.RequestID}))
		}
		return &Output{Body: page}, nil
	})
}

func paramsOf(in *Input) (db.ListEventsParams, error) {
	p := db.ListEventsParams{RowLimit: in.Limit + 1}
	var details []error
	if in.Before > 0 {
		p.Before = pgtype.Int8{Int64: in.Before, Valid: true}
	}
	if in.Action != "" {
		p.Action = pgtype.Text{String: in.Action, Valid: true}
	}
	if in.ActorID != "" {
		id, err := uuid.Parse(in.ActorID)
		if err != nil {
			details = append(details, &huma.ErrorDetail{Location: "query.actor_id", Message: "expected a UUID", Value: in.ActorID})
		}
		p.ActorID = pgtype.UUID{Bytes: id, Valid: err == nil}
	}
	if in.ResourceType != "" {
		p.ResourceType = pgtype.Text{String: in.ResourceType, Valid: true}
	}
	if in.ResourceID != "" {
		if in.ResourceType == "" {
			details = append(details, &huma.ErrorDetail{Location: "query.resource_id", Message: "send resource_type with resource_id", Value: in.ResourceID})
		}
		p.ResourceID = pgtype.Text{String: in.ResourceID, Valid: true}
	}
	for _, bound := range []struct {
		raw, location string
		into          **time.Time
	}{{in.From, "query.from", &p.FromAt}, {in.To, "query.to", &p.ToAt}} {
		if bound.raw == "" {
			continue
		}
		at, err := time.Parse(time.RFC3339Nano, bound.raw)
		if err != nil {
			details = append(details, &huma.ErrorDetail{Location: bound.location, Message: "expected an RFC 3339 timestamp", Value: bound.raw})
			continue
		}
		*bound.into = &at
	}
	if len(details) > 0 {
		return p, huma.Error422UnprocessableEntity("validation failed", details...)
	}
	return p, nil
}
