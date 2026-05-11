package trips

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
	for _, q := range []string{
		`INSERT INTO users(id,email) VALUES('u_a','a@example.com')`,
		`INSERT INTO users(id,email) VALUES('u_b','b@example.com')`,
	} {
		if _, err := d.Exec(q); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	return d
}
