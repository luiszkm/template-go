-- +goose Up
CREATE UNIQUE INDEX roles_name_lower_key ON roles (lower(name));

-- +goose Down
DROP INDEX roles_name_lower_key;
