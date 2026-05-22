package catalog

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/bejl/packing-list/internal/ids"
)

type Bundle struct {
	ID          string
	Name        string
	Description string
	CreatedBy   string
	CreatedAt   time.Time
	DeletedAt   sql.NullTime
	DeletedBy   sql.NullString
}

type BundleItem struct {
	BundleID string
	ItemID   string
	Qty      sql.NullFloat64
}

type Bundles struct{ db *sql.DB }

func NewBundles(db *sql.DB) *Bundles { return &Bundles{db: db} }

func (r *Bundles) Create(ctx context.Context, b Bundle) (string, error) {
	if b.Name == "" {
		return "", ErrValidation
	}
	id := ids.New()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO bundles(id,name,description,created_by) VALUES (?,?,?,?)`,
		id, b.Name, nullStr(b.Description), b.CreatedBy)
	return id, err
}

func (r *Bundles) Get(ctx context.Context, id string) (Bundle, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id,name,COALESCE(description,''),COALESCE(created_by,''),created_at,deleted_at,deleted_by FROM bundles WHERE id = ?`, id)
	var b Bundle
	err := row.Scan(&b.ID, &b.Name, &b.Description, &b.CreatedBy, &b.CreatedAt, &b.DeletedAt, &b.DeletedBy)
	if errors.Is(err, sql.ErrNoRows) {
		return b, ErrNotFound
	}
	return b, err
}

func (r *Bundles) List(ctx context.Context) ([]Bundle, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,name,COALESCE(description,''),COALESCE(created_by,''),created_at,deleted_at,deleted_by FROM bundles WHERE deleted_at IS NULL ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Bundle
	for rows.Next() {
		var b Bundle
		if err := rows.Scan(&b.ID, &b.Name, &b.Description, &b.CreatedBy, &b.CreatedAt, &b.DeletedAt, &b.DeletedBy); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *Bundles) ListDeleted(ctx context.Context) ([]Bundle, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,name,COALESCE(description,''),COALESCE(created_by,''),created_at,deleted_at,deleted_by FROM bundles WHERE deleted_at IS NOT NULL ORDER BY deleted_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Bundle
	for rows.Next() {
		var b Bundle
		if err := rows.Scan(&b.ID, &b.Name, &b.Description, &b.CreatedBy, &b.CreatedAt, &b.DeletedAt, &b.DeletedBy); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *Bundles) Update(ctx context.Context, id string, b Bundle) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE bundles SET name = ?, description = ? WHERE id = ?`,
		b.Name, nullStr(b.Description), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Bundles) SoftDelete(ctx context.Context, id, byUser string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE bundles SET deleted_at = CURRENT_TIMESTAMP, deleted_by = ? WHERE id = ? AND deleted_at IS NULL`,
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

func (r *Bundles) Restore(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE bundles SET deleted_at = NULL, deleted_by = NULL WHERE id = ? AND deleted_at IS NOT NULL`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Bundles) Purge(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, q := range []string{
		`DELETE FROM bundle_items WHERE bundle_id = ?`,
		`DELETE FROM bundle_children WHERE parent_id = ? OR child_id = ?`,
		`DELETE FROM trip_bundles WHERE bundle_id = ?`,
		`DELETE FROM bundles WHERE id = ? AND deleted_at IS NOT NULL`,
	} {
		switch q {
		case `DELETE FROM bundle_children WHERE parent_id = ? OR child_id = ?`:
			if _, err := tx.ExecContext(ctx, q, id, id); err != nil {
				return err
			}
		default:
			if _, err := tx.ExecContext(ctx, q, id); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func (r *Bundles) AddItem(ctx context.Context, bundleID, itemID string, qty *float64) error {
	var q any = nil
	if qty != nil {
		q = *qty
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO bundle_items(bundle_id,item_id,qty) VALUES (?,?,?)
		 ON CONFLICT(bundle_id,item_id) DO UPDATE SET qty = excluded.qty`,
		bundleID, itemID, q)
	return err
}

func (r *Bundles) RemoveItem(ctx context.Context, bundleID, itemID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM bundle_items WHERE bundle_id = ? AND item_id = ?`, bundleID, itemID)
	return err
}

func (r *Bundles) Items(ctx context.Context, bundleID string) ([]BundleItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT bundle_id,item_id,qty FROM bundle_items WHERE bundle_id = ?`, bundleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []BundleItem
	for rows.Next() {
		var bi BundleItem
		if err := rows.Scan(&bi.BundleID, &bi.ItemID, &bi.Qty); err != nil {
			return nil, err
		}
		out = append(out, bi)
	}
	return out, rows.Err()
}

// AddChild nests childID under parentID. Returns ErrConflict if a cycle would form.
func (r *Bundles) AddChild(ctx context.Context, parentID, childID string) error {
	if parentID == childID {
		return errors.New("self-nest forbidden")
	}
	// BFS from childID following bundle_children. If we ever reach parentID, cycle.
	queue := []string{childID}
	seen := map[string]bool{childID: true}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		rows, err := r.db.QueryContext(ctx, `SELECT child_id FROM bundle_children WHERE parent_id = ?`, cur)
		if err != nil {
			return err
		}
		var next []string
		for rows.Next() {
			var c string
			if err := rows.Scan(&c); err != nil {
				rows.Close()
				return err
			}
			next = append(next, c)
		}
		rows.Close()
		for _, c := range next {
			if c == parentID {
				return ErrConflict
			}
			if !seen[c] {
				seen[c] = true
				queue = append(queue, c)
			}
		}
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO bundle_children(parent_id,child_id) VALUES (?,?) ON CONFLICT DO NOTHING`,
		parentID, childID)
	return err
}

func (r *Bundles) RemoveChild(ctx context.Context, parentID, childID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM bundle_children WHERE parent_id = ? AND child_id = ?`, parentID, childID)
	return err
}

func (r *Bundles) Children(ctx context.Context, parentID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT child_id FROM bundle_children WHERE parent_id = ?`, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
