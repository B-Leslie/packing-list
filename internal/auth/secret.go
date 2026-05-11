package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// LoadOrCreateSecret resolves the session secret in this order:
//  1. If envHex is non-empty, decode it (hex) and return.
//  2. Else read $dataDir/.session_secret if it exists.
//  3. Else generate 32 random bytes, write them, return.
func LoadOrCreateSecret(dataDir, envHex string) ([]byte, error) {
	if envHex != "" {
		b, err := hex.DecodeString(envHex)
		if err != nil {
			return nil, errors.New("SESSION_SECRET must be hex")
		}
		if len(b) < 16 {
			return nil, errors.New("SESSION_SECRET too short (need ≥16 bytes)")
		}
		return b, nil
	}
	path := filepath.Join(dataDir, ".session_secret")
	b, err := readFileBytes(path)
	if err == nil {
		decoded, err := hex.DecodeString(string(b))
		if err != nil {
			return nil, errors.New(".session_secret malformed")
		}
		return decoded, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	// Generate + write.
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, []byte(hex.EncodeToString(raw[:])), 0o600); err != nil {
		return nil, err
	}
	return raw[:], nil
}

// readFileBytes is exposed for tests in the same package.
func readFileBytes(path string) ([]byte, error) {
	return os.ReadFile(path)
}
