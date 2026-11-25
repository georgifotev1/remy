package data

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func insertMembership(ctx context.Context, q DBTX, userID, tenantID uuid.UUID, systemRole string) error {
	const query = `
		INSERT INTO memberships (user_id, tenant_id, role_id)
		SELECT $1, $2, id FROM roles
		WHERE tenant_id IS NULL AND is_system AND name = $3`

	tag, err := q.Exec(ctx, query, userID, tenantID, systemRole)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("system role %q not found", systemRole)
	}
	return nil
}
