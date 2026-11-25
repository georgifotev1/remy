package data

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func (m Models) Signup(ctx context.Context, u *User, t *Tenant, s *Session) error {
	return withTx(ctx, m.db, func(tx pgx.Tx) error {
		if err := insertUser(ctx, tx, u); err != nil {
			return err
		}

		if err := insertTenant(ctx, tx, t); err != nil {
			return err
		}

		if err := insertMembership(ctx, tx, u.ID, t.ID, "owner"); err != nil {
			return err
		}

		s.UserID = u.ID
		if err := insertSession(ctx, tx, s); err != nil {
			return err
		}

		return nil
	})
}
