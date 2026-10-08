package password_test

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/argon2"

	"github.com/luiszkm/template-go/internal/features/users/password"
)

var pinnedPHC = regexp.MustCompile(`^\$argon2id\$v=19\$m=65536,t=3,p=4\$[A-Za-z0-9+/]{22}\$[A-Za-z0-9+/]{43}$`)

func TestHash_PinnedPHCFormat(t *testing.T) {
	a, err := password.Hash("senha-longa-123")
	require.NoError(t, err)
	b, err := password.Hash("senha-longa-123")
	require.NoError(t, err)
	require.Regexp(t, pinnedPHC, a)
	require.Regexp(t, pinnedPHC, b)
	require.NotEqual(t, a, b, "a fresh salt per hash")
}

func TestVerify_MatchesOnlyThePassword(t *testing.T) {
	h, err := password.Hash("senha-longa-123")
	require.NoError(t, err)
	ok, err := password.Verify("senha-longa-123", h)
	require.NoError(t, err)
	require.True(t, ok)
	for _, wrong := range []string{"senha-longa-124", "", "SENHA-LONGA-123"} {
		ok, err := password.Verify(wrong, h)
		require.NoError(t, err)
		require.False(t, ok, wrong)
	}
}

func legacyHash(t *testing.T, pw string) string {
	t.Helper()
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	require.NoError(t, err)
	key := argon2.IDKey([]byte(pw), salt, 2, 19456, 1, 32)
	enc := base64.RawStdEncoding
	return fmt.Sprintf("$argon2id$v=19$m=19456,t=2,p=1$%s$%s", enc.EncodeToString(salt), enc.EncodeToString(key))
}

func TestVerify_ReadsParamsFromHash(t *testing.T) {
	old := legacyHash(t, "senha-longa-123")
	ok, err := password.Verify("senha-longa-123", old)
	require.NoError(t, err)
	require.True(t, ok)
	require.True(t, password.NeedsRehash(old))

	current, err := password.Hash("senha-longa-123")
	require.NoError(t, err)
	require.False(t, password.NeedsRehash(current))
}

func TestVerify_RejectsMalformed(t *testing.T) {
	for _, h := range []string{"", "x", "$argon2i$v=19$m=1,t=1,p=1$AAAA$AAAA", "$argon2id$v=19$m=a,t=1,p=1$AAAA$AAAA"} {
		_, err := password.Verify("p", h)
		require.ErrorIs(t, err, password.ErrMalformed, h)
	}
}

func TestDummyHash_IsPinned(t *testing.T) {
	require.Regexp(t, pinnedPHC, password.DummyHash())
}
