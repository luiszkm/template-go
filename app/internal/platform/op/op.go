// Package op is the only way to register an HTTP operation. It refuses an operation that
// declares neither a permission nor the public marker, and a mutation without an audit action.
package op

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

// Permission names what a caller must hold, as "<resource>:<action>", e.g. "users:create".
type Permission string

// Spec declares one operation.
type Spec struct {
	ID            string
	Method        string
	Path          string
	Summary       string
	Tags          []string
	DefaultStatus int

	// Exactly one of Permission or Public must be set.
	Permission Permission
	Public     bool

	// Required for POST, PUT, PATCH and DELETE: the action recorded in the audit log, e.g. "user.created".
	AuditAction string
}

// Metadata keys on huma.Operation, read by the rbac and audit middleware.
const (
	MetaPermission  = "permission"
	MetaPublic      = "public"
	MetaAuditAction = "audit_action"
)

func (s Spec) validate() error {
	switch {
	case s.ID == "":
		return fmt.Errorf("op: operation %s %s has no ID", s.Method, s.Path)
	case s.Permission == "" && !s.Public:
		return fmt.Errorf("op: operation %q declares no permission and is not marked Public", s.ID)
	case s.Permission != "" && s.Public:
		return fmt.Errorf("op: operation %q declares a permission and is marked Public; choose one", s.ID)
	}
	switch s.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		if s.AuditAction == "" {
			return fmt.Errorf("op: mutating operation %q (%s) declares no AuditAction", s.ID, s.Method)
		}
	}
	return nil
}

// Register validates the spec and registers the handler on the API.
func Register[I, O any](api huma.API, s Spec, h func(context.Context, *I) (*O, error)) error {
	if err := s.validate(); err != nil {
		return err
	}
	huma.Register(api, huma.Operation{
		OperationID:   s.ID,
		Method:        s.Method,
		Path:          s.Path,
		Summary:       s.Summary,
		Tags:          s.Tags,
		DefaultStatus: s.DefaultStatus,
		Metadata: map[string]any{
			MetaPermission:  s.Permission,
			MetaPublic:      s.Public,
			MetaAuditAction: s.AuditAction,
		},
	}, h)
	return nil
}
