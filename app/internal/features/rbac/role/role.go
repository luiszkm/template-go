package role

import (
	"strings"
	"unicode/utf8"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/luiszkm/template-go/internal/platform/auth"
	"github.com/luiszkm/template-go/internal/platform/op"
)

const Admin = "admin"

const maxNameLength = 50

type Role struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Permissions []string  `json:"permissions"`
	UserCount   int       `json:"user_count"`
}

type Snapshot struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

func Catalogue(api huma.API) []string {
	items := []string{}
	for _, p := range op.Permissions(api) {
		if p != auth.Wildcard {
			items = append(items, string(p))
		}
	}
	return items
}

func NormalizeName(name string) (string, *huma.ErrorDetail) {
	trimmed := strings.TrimSpace(name)
	if n := utf8.RuneCountInString(trimmed); n < 1 || n > maxNameLength {
		return "", &huma.ErrorDetail{Location: "body.name", Message: "expected 1 to 50 characters after trimming", Value: name}
	}
	return trimmed, nil
}

func CheckPermissions(api huma.API, permissions []string) *huma.ErrorDetail {
	known := map[string]bool{}
	for _, p := range Catalogue(api) {
		known[p] = true
	}
	seen := map[string]bool{}
	for _, p := range permissions {
		switch {
		case !known[p]:
			return &huma.ErrorDetail{Location: "body.permissions", Message: "unknown permission " + p, Value: p}
		case seen[p]:
			return &huma.ErrorDetail{Location: "body.permissions", Message: "duplicate permission " + p, Value: p}
		}
		seen[p] = true
	}
	return nil
}

func Invalid(details ...*huma.ErrorDetail) error {
	errs := make([]error, 0, len(details))
	for _, d := range details {
		if d != nil {
			errs = append(errs, d)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return huma.Error422UnprocessableEntity("validation failed", errs...)
}
