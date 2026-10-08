package getevent_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/features/audit/audittest"
	getevent "github.com/luiszkm/template-go/internal/features/audit/get_event"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

func TestGetEvent_ReturnsDiff(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := audittest.Serve(t, pool, getevent.Register)
	caller := testkit.SignIn(t, pool, "audit:read")
	actor := audittest.InsertUser(t, pool, "a@x.com")
	changed := audittest.Insert(t, pool, audittest.Event{Action: "user.updated", Actor: &actor, ResourceType: "user", ResourceID: "42",
		Before: `{"name":"Ana"}`, After: `{"name":"Bia"}`})
	created := audittest.Insert(t, pool, audittest.Event{Action: "user.created", After: `{"name":"Ana"}`})
	get := func(id int64) map[string]any {
		rec := testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: fmt.Sprintf("/api/v1/audit/events/%d", id), Cookie: caller.Cookie})
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		return testkit.JSON[map[string]any](t, rec)
	}

	body := get(changed)
	require.InDelta(t, float64(changed), body["id"], 0)
	require.Equal(t, "user.updated", body["action"])
	require.Equal(t, map[string]any{"id": actor.String(), "email": "a@x.com"}, body["actor"])
	require.Equal(t, "user", body["resource_type"])
	require.Equal(t, "42", body["resource_id"])
	require.Equal(t, "10.0.0.7", body["ip"])
	require.Equal(t, "req-1", body["request_id"])
	require.Contains(t, body, "occurred_at")
	require.Equal(t, map[string]any{"name": "Ana"}, body["before"])
	require.Equal(t, map[string]any{"name": "Bia"}, body["after"])

	body = get(created)
	require.Contains(t, body, "before")
	require.Nil(t, body["before"])
	require.Equal(t, map[string]any{"name": "Ana"}, body["after"])
}

func TestGetEvent_404And422(t *testing.T) {
	pool := testkit.MigratedDB(t)
	h := audittest.Serve(t, pool, getevent.Register)
	caller := testkit.SignIn(t, pool, "audit:read")
	get := func(id string) int {
		return testkit.Do(t, h, testkit.Request{Method: http.MethodGet, Path: "/api/v1/audit/events/" + id, Cookie: caller.Cookie}).Code
	}
	require.Equal(t, http.StatusNotFound, get("999999"))
	for _, bad := range []string{"abc", "0", "-1"} {
		require.Equal(t, http.StatusUnprocessableEntity, get(bad), bad)
	}
}
