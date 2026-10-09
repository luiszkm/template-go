package listevents_test

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/features/audit/audittest"
	listevents "github.com/luiszkm/template-go/internal/features/audit/list_events"
	"github.com/luiszkm/template-go/internal/platform/testkit"
)

type fixture struct {
	pool   *pgxpool.Pool
	h      http.Handler
	caller testkit.User
}

func setup(t *testing.T) fixture {
	t.Helper()
	pool := testkit.MigratedDB(t)
	return fixture{pool: pool, h: audittest.Serve(t, pool, listevents.Register), caller: testkit.SignIn(t, pool, "audit:read")}
}

type actor struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

type item struct {
	ID           int64  `json:"id"`
	Action       string `json:"action"`
	Actor        *actor `json:"actor"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
}

type page struct {
	Items []item `json:"items"`
	Next  *int64 `json:"next"`
}

func (f fixture) get(t *testing.T, query string) page {
	t.Helper()
	rec := testkit.Do(t, f.h, testkit.Request{Method: http.MethodGet, Path: "/api/v1/audit/events?" + query, Cookie: f.caller.Cookie})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	return testkit.JSON[page](t, rec)
}

func ids(p page) []int64 {
	out := []int64{}
	for _, i := range p.Items {
		out = append(out, i.ID)
	}
	return out
}

func TestListEvents_NewestFirstWithoutDiff(t *testing.T) {
	f := setup(t)
	e1 := audittest.Insert(t, f.pool, audittest.Event{Before: `{"a":1}`, After: `{"a":2}`})
	e2 := audittest.Insert(t, f.pool, audittest.Event{})
	e3 := audittest.Insert(t, f.pool, audittest.Event{})

	require.Equal(t, []int64{e3, e2, e1}, ids(f.get(t, "")))

	rec := testkit.Do(t, f.h, testkit.Request{Method: http.MethodGet, Path: "/api/v1/audit/events", Cookie: f.caller.Cookie})
	raw := testkit.JSON[struct {
		Items []map[string]any `json:"items"`
	}](t, rec)
	for _, it := range raw.Items {
		keys := make([]string, 0, len(it))
		for k := range it {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		require.Equal(t, []string{"action", "actor", "id", "ip", "occurred_at", "request_id", "resource_id", "resource_type"}, keys)
	}
}

func TestListEvents_Actor(t *testing.T) {
	f := setup(t)
	a := audittest.InsertUser(t, f.pool, "a@x.com")
	byA := audittest.Insert(t, f.pool, audittest.Event{Actor: &a})
	bySystem := audittest.Insert(t, f.pool, audittest.Event{})

	actors := func() map[int64]*actor {
		out := map[int64]*actor{}
		for _, i := range f.get(t, "").Items {
			out[i.ID] = i.Actor
		}
		return out
	}
	got := actors()
	require.Equal(t, &actor{ID: a, Email: "a@x.com"}, got[byA])
	require.Nil(t, got[bySystem])

	_, err := f.pool.Exec(t.Context(), `UPDATE users SET email = 'b@x.com' WHERE id = $1`, a)
	require.NoError(t, err)
	require.Equal(t, "b@x.com", actors()[byA].Email)
}

func TestListEvents_Pages(t *testing.T) {
	f := setup(t)
	var all []int64
	for range 51 {
		all = append(all, audittest.Insert(t, f.pool, audittest.Event{}))
	}
	slices.Reverse(all)

	p := f.get(t, "")
	require.Len(t, p.Items, 50)
	require.NotNil(t, p.Next)
	require.Equal(t, all[49], *p.Next)

	g := setup(t)
	for range 3 {
		audittest.Insert(t, g.pool, audittest.Event{})
	}
	two := g.get(t, "limit=2")
	require.Len(t, two.Items, 2)
	require.NotNil(t, two.Next)
	require.Equal(t, two.Items[1].ID, *two.Next)
	require.Nil(t, g.get(t, "limit=3").Next)
}

func TestListEvents_Before(t *testing.T) {
	f := setup(t)
	e1 := audittest.Insert(t, f.pool, audittest.Event{})
	e2 := audittest.Insert(t, f.pool, audittest.Event{})
	e3 := audittest.Insert(t, f.pool, audittest.Event{})

	require.Equal(t, []int64{e2, e1}, ids(f.get(t, fmt.Sprintf("before=%d", e3))))
	last := f.get(t, fmt.Sprintf("before=%d", e1))
	require.Empty(t, last.Items)
	require.Nil(t, last.Next)
}

func TestListEvents_Validation(t *testing.T) {
	f := setup(t)
	rejected := []struct{ query, location string }{
		{"limit=0", "query.limit"},
		{"limit=101", "query.limit"},
		{"before=0", "query.before"},
		{"from=ontem", "query.from"},
		{"to=ontem", "query.to"},
		{"actor_id=abc", "query.actor_id"},
		{"resource_id=x", "query.resource_id"},
	}
	for _, c := range rejected {
		rec := testkit.Do(t, f.h, testkit.Request{Method: http.MethodGet, Path: "/api/v1/audit/events?" + c.query, Cookie: f.caller.Cookie})
		require.Equal(t, http.StatusUnprocessableEntity, rec.Code, c.query)
		require.Contains(t, rec.Body.String(), `"location":"`+c.location+`"`, c.query)
	}
	for _, q := range []string{"limit=1", "limit=100"} {
		f.get(t, q)
	}
}

func TestListEvents_FilterAction(t *testing.T) {
	f := setup(t)
	c1 := audittest.Insert(t, f.pool, audittest.Event{Action: "user.created"})
	audittest.Insert(t, f.pool, audittest.Event{Action: "user.updated"})
	c2 := audittest.Insert(t, f.pool, audittest.Event{Action: "user.created"})
	require.Equal(t, []int64{c2, c1}, ids(f.get(t, "action=user.created")))
}

func TestListEvents_FilterActor(t *testing.T) {
	f := setup(t)
	a := audittest.InsertUser(t, f.pool, "a@x.com")
	b := audittest.InsertUser(t, f.pool, "b@x.com")
	byA := audittest.Insert(t, f.pool, audittest.Event{Actor: &a})
	audittest.Insert(t, f.pool, audittest.Event{Actor: &b})
	audittest.Insert(t, f.pool, audittest.Event{})
	require.Equal(t, []int64{byA}, ids(f.get(t, "actor_id="+a.String())))
}

func TestListEvents_FilterResource(t *testing.T) {
	f := setup(t)
	u1 := audittest.Insert(t, f.pool, audittest.Event{ResourceType: "user", ResourceID: "1"})
	u2 := audittest.Insert(t, f.pool, audittest.Event{ResourceType: "user", ResourceID: "2"})
	audittest.Insert(t, f.pool, audittest.Event{ResourceType: "role", ResourceID: "1"})
	require.Equal(t, []int64{u2, u1}, ids(f.get(t, "resource_type=user")))
	require.Equal(t, []int64{u1}, ids(f.get(t, "resource_type=user&resource_id=1")))
}

func TestListEvents_FilterPeriod(t *testing.T) {
	f := setup(t)
	day := time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC)
	at := func(h int) time.Time { return day.Add(time.Duration(h) * time.Hour) }
	ten := audittest.Insert(t, f.pool, audittest.Event{At: at(10)})
	eleven := audittest.Insert(t, f.pool, audittest.Event{At: at(11)})
	twelve := audittest.Insert(t, f.pool, audittest.Event{At: at(12)})
	q := func(k string, h int) string { return k + "=" + url.QueryEscape(at(h).Format(time.RFC3339)) }

	require.Equal(t, []int64{twelve, eleven}, ids(f.get(t, q("from", 11))))
	require.Equal(t, []int64{eleven, ten}, ids(f.get(t, q("to", 12))))
	require.Equal(t, []int64{eleven}, ids(f.get(t, strings.Join([]string{q("from", 11), q("to", 12)}, "&"))))
}

func TestListEvents_FiltersCombineAndPage(t *testing.T) {
	f := setup(t)
	var created []int64
	for range 3 {
		created = append(created, audittest.Insert(t, f.pool, audittest.Event{Action: "user.created"}))
		audittest.Insert(t, f.pool, audittest.Event{Action: "user.updated"})
	}
	first := f.get(t, "action=user.created&limit=2")
	require.Equal(t, []int64{created[2], created[1]}, ids(first))
	require.NotNil(t, first.Next)
	second := f.get(t, fmt.Sprintf("action=user.created&limit=2&before=%d", *first.Next))
	require.Equal(t, []int64{created[0]}, ids(second))
	require.Nil(t, second.Next)
}

func TestListEvents_IPPresentOrNull(t *testing.T) {
	f := setup(t)
	withIP := audittest.Insert(t, f.pool, audittest.Event{})
	withoutIP := audittest.Insert(t, f.pool, audittest.Event{NoIP: true})
	rec := testkit.Do(t, f.h, testkit.Request{Method: http.MethodGet, Path: "/api/v1/audit/events", Cookie: f.caller.Cookie})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	ips := map[int64]*string{}
	for _, it := range testkit.JSON[struct {
		Items []struct {
			ID int64   `json:"id"`
			IP *string `json:"ip"`
		} `json:"items"`
	}](t, rec).Items {
		ips[it.ID] = it.IP
	}
	require.NotNil(t, ips[withIP])
	require.Equal(t, "10.0.0.7", *ips[withIP])
	require.Contains(t, ips, withoutIP)
	require.Nil(t, ips[withoutIP])
	require.Contains(t, rec.Body.String(), `"ip":null`)
}
