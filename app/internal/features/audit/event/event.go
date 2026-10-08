package event

import (
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuditActor struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

func (AuditActor) Schema(huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type:     huma.TypeObject,
		Nullable: true,
		Required: []string{"id", "email"},
		Properties: map[string]*huma.Schema{
			"id":    {Type: huma.TypeString, Format: "uuid"},
			"email": {Type: huma.TypeString},
		},
		AdditionalProperties: false,
	}
}

type AuditEvent struct {
	ID           int64       `json:"id"`
	OccurredAt   time.Time   `json:"occurred_at"`
	Action       string      `json:"action"`
	Actor        *AuditActor `json:"actor"`
	ResourceType string      `json:"resource_type"`
	ResourceID   string      `json:"resource_id"`
	IP           *string     `json:"ip" nullable:"true"`
	RequestID    string      `json:"request_id"`
}

type Row struct {
	ID           int64
	OccurredAt   time.Time
	Action       string
	ActorID      pgtype.UUID
	ActorEmail   pgtype.Text
	ResourceType string
	ResourceID   string
	IP           string
	RequestID    string
}

func From(r Row) AuditEvent {
	e := AuditEvent{ID: r.ID, OccurredAt: r.OccurredAt, Action: r.Action, ResourceType: r.ResourceType,
		ResourceID: r.ResourceID, RequestID: r.RequestID}
	if r.ActorID.Valid {
		e.Actor = &AuditActor{ID: r.ActorID.Bytes, Email: r.ActorEmail.String}
	}
	if r.IP != "" {
		e.IP = &r.IP
	}
	return e
}
