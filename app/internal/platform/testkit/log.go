package testkit

import (
	"log/slog"
)

func DiscardLogger() *slog.Logger { return slog.New(slog.DiscardHandler) }
