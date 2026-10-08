package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/crypto/argon2"
)

const (
	memoryKiB = 64 * 1024
	passes    = 3
	threads   = 4
	saltLen   = 16
	keyLen    = 32
)

const (
	MinLen = 12
	MaxLen = 128
)

var b64 = base64.RawStdEncoding

var slots = make(chan struct{}, max(2, runtime.NumCPU()))

type params struct {
	memory  uint32
	time    uint32
	threads uint8
}

var pinned = params{memory: memoryKiB, time: passes, threads: threads}

func Hash(pw string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("password: salt: %w", err)
	}
	key := derive(pw, salt, pinned, keyLen)
	return encode(pinned, salt, key), nil
}

func encode(p params, salt, key []byte) string {
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.memory, p.time, p.threads, b64.EncodeToString(salt), b64.EncodeToString(key))
}

func derive(pw string, salt []byte, p params, n uint32) []byte {
	slots <- struct{}{}
	defer func() { <-slots }()
	return argon2.IDKey([]byte(pw), salt, p.time, p.memory, p.threads, n)
}

var ErrMalformed = errors.New("password: malformed hash")

func decode(hash string) (params, []byte, []byte, error) {
	parts := strings.Split(hash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return params{}, nil, nil, ErrMalformed
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return params{}, nil, nil, ErrMalformed
	}
	var p params
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.memory, &p.time, &p.threads); err != nil || p.time == 0 || p.threads == 0 {
		return params{}, nil, nil, ErrMalformed
	}
	salt, err := b64.DecodeString(parts[4])
	if err != nil {
		return params{}, nil, nil, ErrMalformed
	}
	key, err := b64.DecodeString(parts[5])
	if err != nil || len(key) == 0 {
		return params{}, nil, nil, ErrMalformed
	}
	return p, salt, key, nil
}

func Verify(pw, hash string) (bool, error) {
	p, salt, key, err := decode(hash)
	if err != nil {
		return false, err
	}
	got := derive(pw, salt, p, uint32(len(key))) //nolint:gosec
	return subtle.ConstantTimeCompare(got, key) == 1, nil
}

func NeedsRehash(hash string) bool {
	p, _, key, err := decode(hash)
	return err != nil || p != pinned || len(key) != keyLen
}

type Verifier func(pw, hash string) (bool, error)

var dummyHash = sync.OnceValue(func() string {
	h, err := Hash("not-a-real-password-for-timing-only")
	if err != nil {
		panic(err)
	}
	return h
})

func DummyHash() string { return dummyHash() }
