// Package catalog holds the global item + bundle catalog.
package catalog

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/bejl/packing-list/internal/ids"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("conflict")
	ErrValidation = errors.New("validation")
)

type Item struct {
	ID         string
	Name       string
	Category   string
	PerNight   bool
	DefaultQty int
	Notes      string
	CreatedBy  string
	CreatedAt  time.Time
	DeletedAt  sql.NullTime
	DeletedBy  sql.NullString
}

type Items struct{ db *sql.DB }

func NewItems(db *sql.DB) *Items { return &Items{db: db} }

func (r *Items) Create(ctx context.Context, it Item) (string, error) {
	if it.Name == "" {
		return "", ErrValidation
	}
	if it.Category == "" {
		it.Category = "general"
	}
	if it.DefaultQty <= 0 {
		it.DefaultQty = 1
	}
	id := ids.New()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO items(id,name,category,per_night,default_qty,notes,created_by) VALUES (?,?,?,?,?,?,?)`,
		id, it.Name, it.Category, boolInt(it.PerNight), it.DefaultQty, nullStr(it.Notes), it.CreatedBy)
	return id, err
}

func (r *Items) Get(ctx context.Context, id string) (Item, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id,name,category,per_night,default_qty,COALESCE(notes,''),COALESCE(created_by,''),created_at,deleted_at,deleted_by
		 FROM items WHERE id = ?`, id)
	return scanItem(row)
}

func (r *Items) List(ctx context.Context) ([]Item, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,name,category,per_night,default_qty,COALESCE(notes,''),COALESCE(created_by,''),created_at,deleted_at,deleted_by
		 FROM items WHERE deleted_at IS NULL ORDER BY category, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Item
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *Items) ListDeleted(ctx context.Context) ([]Item, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,name,category,per_night,default_qty,COALESCE(notes,''),COALESCE(created_by,''),created_at,deleted_at,deleted_by
		 FROM items WHERE deleted_at IS NOT NULL ORDER BY deleted_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Item
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *Items) Update(ctx context.Context, id string, it Item) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE items SET name = ?, category = ?, per_night = ?, default_qty = ?, notes = ? WHERE id = ?`,
		it.Name, it.Category, boolInt(it.PerNight), it.DefaultQty, nullStr(it.Notes), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Items) SoftDelete(ctx context.Context, id, byUser string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE items SET deleted_at = CURRENT_TIMESTAMP, deleted_by = ? WHERE id = ? AND deleted_at IS NULL`,
		byUser, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Items) Restore(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE items SET deleted_at = NULL, deleted_by = NULL WHERE id = ? AND deleted_at IS NOT NULL`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Items) Purge(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, q := range []string{
		`DELETE FROM bundle_items WHERE item_id = ?`,
		`DELETE FROM trip_extras WHERE item_id = ?`,
		`DELETE FROM trip_overrides WHERE item_id = ?`,
		`DELETE FROM trip_pack_state WHERE item_id = ?`,
		`DELETE FROM items WHERE id = ? AND deleted_at IS NOT NULL`,
	} {
		if _, err := tx.ExecContext(ctx, q, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

type rowScanner interface {
	Scan(...any) error
}

func scanItem(s rowScanner) (Item, error) {
	var it Item
	var per int
	err := s.Scan(&it.ID, &it.Name, &it.Category, &per, &it.DefaultQty, &it.Notes,
		&it.CreatedBy, &it.CreatedAt, &it.DeletedAt, &it.DeletedBy)
	if errors.Is(err, sql.ErrNoRows) {
		return it, ErrNotFound
	}
	it.PerNight = per != 0
	return it, err
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
