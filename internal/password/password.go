package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	memory     uint32 = 64 * 1024 // KiB = 64 MiB
	iterations uint32 = 1
	threads    uint8  = 2
	saltLen           = 16
	keyLen     uint32 = 32
)

var ErrInvalidHash = errors.New("password: invalid hash format")

var b64 = base64.RawStdEncoding

func Hash(plain string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	key := argon2.IDKey([]byte(plain), salt, iterations, memory, threads, keyLen)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		memory,
		iterations,
		threads,
		b64.EncodeToString(salt),
		b64.EncodeToString(key),
	), nil
}

// Verify reports whether plain matches the encoded hash.
func Verify(plain, encoded string) (bool, error) {
	// "", "argon2id", "v=19", "m=...,t=...,p=...", salt, key
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, ErrInvalidHash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return false, ErrInvalidHash
	}

	var (
		m uint32
		t uint32
		p uint8
	)
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p); err != nil {
		return false, ErrInvalidHash
	}

	salt, err := b64.DecodeString(parts[4])
	if err != nil {
		return false, ErrInvalidHash
	}
	want, err := b64.DecodeString(parts[5])
	if err != nil {
		return false, ErrInvalidHash
	}

	got := argon2.IDKey([]byte(plain), salt, t, m, p, uint32(len(want)))

	return subtle.ConstantTimeCompare(got, want) == 1, nil
}
