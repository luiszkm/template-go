package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/luiszkm/template-go/internal/platform/auth"
	"github.com/luiszkm/template-go/internal/platform/httpx"
)

type Event struct {
	Action       string
	ResourceType string
	ResourceID   string
	Before       any
	After        any
}

var redacted = map[string]bool{"password_hash": true, "password": true, "current_password": true, "new_password": true}

func Record(ctx context.Context, tx pgx.Tx, e Event) error {
	before, err := redactedJSON(e.Before)
	if err != nil {
		return err
	}
	after, err := redactedJSON(e.After)
	if err != nil {
		return err
	}
	var actor any
	if p, ok := auth.PrincipalFrom(ctx); ok {
		actor = p.UserID
	}
	var ip any
	if addr := httpx.ClientIP(ctx); addr.IsValid() {
		ip = addr.String()
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO audit_events (actor_id, action, resource_type, resource_id, before, after, ip, request_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		actor, e.Action, e.ResourceType, e.ResourceID, before, after, ip, httpx.RequestIDFrom(ctx))
	if err != nil {
		return fmt.Errorf("audit: record %s: %w", e.Action, err)
	}
	return nil
}

func redactedJSON(v any) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("audit: marshal: %w", err)
	}
	var tree any
	if err := json.Unmarshal(raw, &tree); err != nil {
		return nil, fmt.Errorf("audit: unmarshal: %w", err)
	}
	return json.Marshal(redact(tree))
}

func redact(v any) any {
	switch t := v.(type) {
	case map[string]any:
		for k, child := range t {
			if redacted[k] {
				delete(t, k)
				continue
			}
			t[k] = redact(child)
		}
	case []any:
		for i, child := range t {
			t[i] = redact(child)
		}
	}
	return v
}
