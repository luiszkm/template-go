package me

import (
	"context"
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/luiszkm/template-go/internal/features/users/me/db"
	"github.com/luiszkm/template-go/internal/platform/auth"
	"github.com/luiszkm/template-go/internal/platform/deps"
	"github.com/luiszkm/template-go/internal/platform/op"
)

type Input struct{}

type Me struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	Permissions []string  `json:"permissions"`
}

type Output struct {
	Body Me
}

func Register(api huma.API, d deps.Deps) error {
	return op.Register(api, op.Spec{
		ID:            "get-me",
		Method:        http.MethodGet,
		Path:          "/api/v1/users/me",
		Summary:       "The signed-in user and the permissions they hold",
		Tags:          []string{"users"},
		Authenticated: true,
	}, func(ctx context.Context, _ *Input) (*Output, error) {
		principal, _ := auth.PrincipalFrom(ctx)
		user, err := db.New(d.DB).UserByID(ctx, principal.UserID)
		if err != nil {
			return nil, err
		}
		return &Output{Body: Me{ID: user.ID, Email: user.Email, Name: user.Name, Permissions: heldPermissions(api, principal)}}, nil
	})
}

func heldPermissions(api huma.API, p auth.Principal) []string {
	held := []string{}
	if p.Can(auth.Wildcard) {
		for _, perm := range op.Permissions(api) {
			held = append(held, string(perm))
		}
		return held
	}
	for perm := range p.Permissions {
		held = append(held, string(perm))
	}
	slices.Sort(held)
	return held
}
