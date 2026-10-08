package email

import (
	"net/mail"
	"strings"
)

const MaxLen = 254

func Normalize(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

func Valid(s string) bool {
	if s == "" || len(s) > MaxLen {
		return false
	}
	addr, err := mail.ParseAddress(s)
	return err == nil && addr.Name == "" && addr.Address == s && strings.Contains(s, "@")
}
