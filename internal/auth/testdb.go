package auth

import (
	"database/sql"
	"path/filepath"
	"testing"

	pdb "github.com/bejl/packing-list/internal/db"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	d, err := pdb.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}
