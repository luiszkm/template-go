package testkit

import (
	"log/slog"
)

// DiscardLogger is a logger for tests that do not assert on logs.
func DiscardLogger() *slog.Logger { return slog.New(slog.DiscardHandler) }
