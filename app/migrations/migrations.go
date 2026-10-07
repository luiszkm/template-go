// Package migrations embeds the goose SQL migrations. Name files YYYYMMDDHHMMSS_<name>.sql.
package migrations

import "embed"

// FS holds every *.sql file in this directory.
//
//go:embed *.sql
var FS embed.FS
