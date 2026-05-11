// Package ids generates ULIDs for primary keys.
package ids

import (
	"crypto/rand"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	mu     sync.Mutex
	source = ulid.Monotonic(rand.Reader, 0)
)

// New returns a fresh ULID as a 26-char base32 string.
func New() string {
	mu.Lock()
	defer mu.Unlock()
	return ulid.MustNew(ulid.Timestamp(time.Now()), source).String()
}
