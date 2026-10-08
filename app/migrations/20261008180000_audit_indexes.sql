-- +goose Up
CREATE INDEX audit_events_actor_idx ON audit_events (actor_id, id DESC);
CREATE INDEX audit_events_resource_idx ON audit_events (resource_type, resource_id, id DESC);
CREATE INDEX audit_events_action_idx ON audit_events (action, id DESC);
CREATE INDEX audit_events_occurred_idx ON audit_events (occurred_at);

-- +goose Down
DROP INDEX audit_events_occurred_idx;
DROP INDEX audit_events_action_idx;
DROP INDEX audit_events_resource_idx;
DROP INDEX audit_events_actor_idx;
