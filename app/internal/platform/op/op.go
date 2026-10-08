package op

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"
)

type Permission string

type Spec struct {
	ID            string
	Method        string
	Path          string
	Summary       string
	Tags          []string
	DefaultStatus int
	Errors        []int

	Permission    Permission
	Public        bool
	Authenticated bool

	AuditAction string
}

const (
	MetaPermission    = "permission"
	MetaPublic        = "public"
	MetaAuthenticated = "authenticated"
	MetaAuditAction   = "audit_action"
)

func (s Spec) validate() error {
	if s.ID == "" {
		return fmt.Errorf("op: operation %s %s has no ID", s.Method, s.Path)
	}
	markers := 0
	for _, set := range []bool{s.Permission != "", s.Public, s.Authenticated} {
		if set {
			markers++
		}
	}
	switch markers {
	case 0:
		return fmt.Errorf("op: operation %q declares no permission and is not marked Public or Authenticated", s.ID)
	case 1:
	default:
		return fmt.Errorf("op: operation %q declares more than one of Permission, Public and Authenticated; choose one", s.ID)
	}
	switch s.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		if s.AuditAction == "" {
			return fmt.Errorf("op: mutating operation %q (%s) declares no AuditAction", s.ID, s.Method)
		}
	}
	return nil
}

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
		Errors:        s.documentedErrors(),
		Metadata: map[string]any{
			MetaPermission:    s.Permission,
			MetaPublic:        s.Public,
			MetaAuthenticated: s.Authenticated,
			MetaAuditAction:   s.AuditAction,
		},
	}, h)
	return nil
}

func Permissions(api huma.API) []Permission {
	var out []Permission
	for _, item := range api.OpenAPI().Paths {
		for _, o := range []*huma.Operation{item.Get, item.Put, item.Post, item.Delete, item.Options, item.Head, item.Patch, item.Trace} {
			if o == nil {
				continue
			}
			if p, ok := o.Metadata[MetaPermission].(Permission); ok && p != "" {
				out = append(out, p)
			}
		}
	}
	slices.Sort(out)
	return slices.Compact(out)
}

func (s Spec) documentedErrors() []int {
	errs := slices.Clone(s.Errors)
	if !s.Public {
		errs = append(errs, http.StatusUnauthorized)
	}
	if s.Permission != "" {
		errs = append(errs, http.StatusForbidden)
	}
	slices.Sort(errs)
	return slices.Compact(errs)
}
