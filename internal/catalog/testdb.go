package catalog

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
	// Seed a user for created_by FK.
	if _, err := d.Exec(`INSERT INTO users(id,email) VALUES('u_test','test@example.com')`); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return d
}
