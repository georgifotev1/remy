package data

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// insertTenant inserts t and fills in its ID and timestamps.
func insertTenant(ctx context.Context, q DBTX, t *Tenant) error {
	const query = `
		INSERT INTO tenants (name, slug)
		VALUES ($1, $2)
		RETURNING id, name, slug, created_at, updated_at`

	err := q.QueryRow(ctx, query, t.Name, t.Slug).Scan(
		&t.ID,
		&t.Name,
		&t.Slug,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if isUniqueViolation(err, "tenants_slug_uq") {
		return ErrDuplicateSlug
	}
	return err
}
