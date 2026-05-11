// Package db opens SQLite with the right pragmas and applies bundled migrations.
package db

import (
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Open opens (or creates) the SQLite database at path, applies any pending
// migrations, and returns a *sql.DB ready for use.
func Open(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)", path)
	d, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	if err := d.Ping(); err != nil {
		d.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	if err := migrate(d); err != nil {
		d.Close()
		return nil, err
	}
	return d, nil
}

func migrate(d *sql.DB) error {
	if _, err := d.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	type m struct {
		v   int
		sql string
	}
	var ms []m
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		// File name format: NNNN_*.sql
		idx := strings.Index(e.Name(), "_")
		if idx < 0 {
			continue
		}
		v, err := strconv.Atoi(e.Name()[:idx])
		if err != nil {
			return fmt.Errorf("bad migration name %q: %w", e.Name(), err)
		}
		body, err := migrationsFS.ReadFile("migrations/" + e.Name())
		if err != nil {
			return err
		}
		ms = append(ms, m{v, string(body)})
	}
	sort.Slice(ms, func(i, j int) bool { return ms[i].v < ms[j].v })

	for _, mig := range ms {
		var present int
		if err := d.QueryRow(`SELECT count(*) FROM schema_migrations WHERE version = ?`, mig.v).Scan(&present); err != nil {
			return err
		}
		if present > 0 {
			continue
		}
		tx, err := d.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(mig.sql); err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration %d: %w", mig.v, err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations(version) VALUES (?)`, mig.v); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
