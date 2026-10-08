package bootstrap_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luiszkm/template-go/internal/features/users/bootstrap"
)

func TestCreateAdmin_PasswordBounds(t *testing.T) {
	for _, n := range []int{12, 128} {
		require.NoError(t, bootstrap.Validate("a@x.com", "Ana", strings.Repeat("p", n)), n)
	}
	for _, n := range []int{11, 129} {
		require.ErrorIs(t, bootstrap.Validate("a@x.com", "Ana", strings.Repeat("p", n)), bootstrap.ErrInvalid, n)
	}
	require.ErrorIs(t, bootstrap.Validate("", "Ana", strings.Repeat("p", 12)), bootstrap.ErrInvalid)
	require.ErrorIs(t, bootstrap.Validate("a@x.com", "", strings.Repeat("p", 12)), bootstrap.ErrInvalid)
}
