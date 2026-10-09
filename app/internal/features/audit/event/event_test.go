package event_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/features/audit/event"
)

func TestFrom_ActorAndIPArms(t *testing.T) {
	actor := uuid.New()
	ip := "10.0.0.7"
	cases := []struct {
		name      string
		row       event.Row
		wantActor *event.AuditActor
		wantIP    *string
	}{
		{"actor present", event.Row{ActorID: pgtype.UUID{Bytes: actor, Valid: true}, ActorEmail: pgtype.Text{String: "a@x.com", Valid: true}, IP: ip},
			&event.AuditActor{ID: actor, Email: "a@x.com"}, &ip},
		{"actor null", event.Row{IP: ip}, nil, &ip},
		{"ip present", event.Row{IP: ip}, nil, &ip},
		{"ip null", event.Row{IP: ""}, nil, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.row.ID, c.row.OccurredAt, c.row.Action, c.row.ResourceType, c.row.ResourceID, c.row.RequestID =
				7, time.Unix(0, 0), "user.created", "user", "42", "req-1"
			got := event.From(c.row)
			require.Equal(t, c.wantActor, got.Actor)
			require.Equal(t, c.wantIP, got.IP)
			require.Equal(t, int64(7), got.ID)
			require.Equal(t, "user.created", got.Action)
			require.Equal(t, "req-1", got.RequestID)
		})
	}
}
