// Package trips owns trips, trip membership, and trip-scoped sub-tables.
package trips

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/bejl/packing-list/internal/ids"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrNotMember = errors.New("not a member")
	ErrForbidden = errors.New("forbidden")
)

type Trip struct {
	ID        string
	Name      string
	Nights    int
	StartsOn  sql.NullString
	Notes     string
	OwnerID   string
	CreatedAt time.Time
	DeletedAt sql.NullTime
	DeletedBy sql.NullString
}

type Member struct {
	UserID  string
	Role    string
	Email   string
	AddedAt time.Time
}

type Trips struct{ db *sql.DB }

func NewTrips(db *sql.DB) *Trips { return &Trips{db: db} }

func (r *Trips) Create(ctx context.Context, name string, nights int, ownerID string) (string, error) {
	if name == "" || nights < 0 {
		return "", errors.New("invalid trip")
	}
	id := ids.New()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO trips(id,name,nights,owner_id) VALUES (?,?,?,?)`,
		id, name, nights, ownerID); err != nil {
		return "", err
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO trip_members(trip_id,user_id,role) VALUES (?,?,'owner')`,
		id, ownerID); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return id, nil
}

func (r *Trips) Get(ctx context.Context, id string) (Trip, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id,name,nights,starts_on,COALESCE(notes,''),owner_id,created_at,deleted_at,deleted_by
		 FROM trips WHERE id = ?`, id)
	var t Trip
	err := row.Scan(&t.ID, &t.Name, &t.Nights, &t.StartsOn, &t.Notes, &t.OwnerID, &t.CreatedAt, &t.DeletedAt, &t.DeletedBy)
	if errors.Is(err, sql.ErrNoRows) {
		return t, ErrNotFound
	}
	return t, err
}

func (r *Trips) Update(ctx context.Context, id, name string, nights int, notes string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE trips SET name = ?, nights = ?, notes = ? WHERE id = ?`,
		name, nights, nullStr(notes), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Trips) SoftDelete(ctx context.Context, id, byUser string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE trips SET deleted_at = CURRENT_TIMESTAMP, deleted_by = ? WHERE id = ? AND deleted_at IS NULL`,
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

func (r *Trips) Restore(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE trips SET deleted_at = NULL, deleted_by = NULL WHERE id = ? AND deleted_at IS NOT NULL`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Trips) Purge(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, q := range []string{
		`DELETE FROM trip_pack_state WHERE trip_id = ?`,
		`DELETE FROM trip_overrides WHERE trip_id = ?`,
		`DELETE FROM trip_extras WHERE trip_id = ?`,
		`DELETE FROM trip_bundles WHERE trip_id = ?`,
		`DELETE FROM trip_members WHERE trip_id = ?`,
		`DELETE FROM trips WHERE id = ? AND deleted_at IS NOT NULL`,
	} {
		if _, err := tx.ExecContext(ctx, q, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Trips) ListVisibleTo(ctx context.Context, userID string) ([]Trip, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id,t.name,t.nights,t.starts_on,COALESCE(t.notes,''),t.owner_id,t.created_at,t.deleted_at,t.deleted_by
		FROM trips t
		JOIN trip_members m ON m.trip_id = t.id
		WHERE t.deleted_at IS NULL AND m.user_id = ?
		ORDER BY t.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Trip
	for rows.Next() {
		var t Trip
		if err := rows.Scan(&t.ID, &t.Name, &t.Nights, &t.StartsOn, &t.Notes, &t.OwnerID, &t.CreatedAt, &t.DeletedAt, &t.DeletedBy); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *Trips) ListDeleted(ctx context.Context, userID string) ([]Trip, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id,t.name,t.nights,t.starts_on,COALESCE(t.notes,''),t.owner_id,t.created_at,t.deleted_at,t.deleted_by
		FROM trips t
		JOIN trip_members m ON m.trip_id = t.id
		WHERE t.deleted_at IS NOT NULL AND m.user_id = ?
		ORDER BY t.deleted_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Trip
	for rows.Next() {
		var t Trip
		if err := rows.Scan(&t.ID, &t.Name, &t.Nights, &t.StartsOn, &t.Notes, &t.OwnerID, &t.CreatedAt, &t.DeletedAt, &t.DeletedBy); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *Trips) AddMember(ctx context.Context, tripID, userID, role string) error {
	if role != "owner" && role != "editor" {
		return errors.New("bad role")
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO trip_members(trip_id,user_id,role) VALUES (?,?,?) ON CONFLICT DO NOTHING`,
		tripID, userID, role)
	return err
}

func (r *Trips) RemoveMember(ctx context.Context, tripID, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM trip_members WHERE trip_id = ? AND user_id = ? AND role <> 'owner'`, tripID, userID)
	return err
}

func (r *Trips) Members(ctx context.Context, tripID string) ([]Member, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT m.user_id, m.role, u.email, m.added_at
		FROM trip_members m JOIN users u ON u.id = m.user_id
		WHERE m.trip_id = ?
		ORDER BY (m.role = 'owner') DESC, m.added_at`, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Member
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.UserID, &m.Role, &m.Email, &m.AddedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *Trips) RoleOf(ctx context.Context, tripID, userID string) (string, error) {
	var role string
	err := r.db.QueryRowContext(ctx,
		`SELECT role FROM trip_members WHERE trip_id = ? AND user_id = ?`, tripID, userID).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotMember
	}
	return role, err
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
