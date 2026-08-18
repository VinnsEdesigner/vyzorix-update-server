package storage

import (
	"context"
	"database/sql"
	"errors"
)

// migrateOrganizations creates the organizations, organization_members, and invitations tables.
// This migration implements the multi-tenant organization model where:.
// - Operators are global identities (email, password, OAuth).
// - Organizations are tenants that own resources.
// - Memberships link operators to organizations with scoped roles.
// - Invitations manage the join request flow.
func migrateOrganizations(tx *sql.Tx) error {
	// Create organizations table.
	if err := createOrganizationsTable(tx); err != nil {
		return err
	}

	// Create organization_members table.
	if err := createOrganizationMembersTable(tx); err != nil {
		return err
	}

	// Create invitations table.
	if err := createInvitationsTable(tx); err != nil {
		return err
	}

	// Add organization_id to devices table.
	if err := addOrganizationIDToDevices(tx); err != nil {
		return err
	}

	// Add organization_id to sessions table.
	if err := addOrganizationIDToSessions(tx); err != nil {
		return err
	}

	// Add organization_id to api_keys table.
	if err := addOrganizationIDToAPIKeys(tx); err != nil {
		return err
	}

	return nil
}

// createOrganizationsTable creates the organizations table.
func createOrganizationsTable(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS organizations (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			created_by TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			deleted_at INTEGER,
			is_active INTEGER NOT NULL DEFAULT 1,
			max_members INTEGER NOT NULL DEFAULT 2,
			FOREIGN KEY (created_by) REFERENCES operators(id),
			UNIQUE(created_by, name)
		)
	`)
	if err != nil {
		return err
	}

	// Create index for looking up organizations by creator.
	_, err = tx.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_organizations_created_by ON organizations(created_by)
	`)

	return err
}

// createOrganizationMembersTable creates the organization_members table.
func createOrganizationMembersTable(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS organization_members (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL,
			operator_id TEXT NOT NULL,
			role TEXT NOT NULL CHECK (role IN ('super_admin', 'admin', 'operator', 'viewer')),
			invited_by TEXT,
			joined_at INTEGER NOT NULL,
			removed_at INTEGER,
			status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'removed')),
			UNIQUE(organization_id, operator_id),
			FOREIGN KEY (organization_id) REFERENCES organizations(id),
			FOREIGN KEY (operator_id) REFERENCES operators(id),
			FOREIGN KEY (invited_by) REFERENCES operators(id)
		)
	`)
	if err != nil {
		return err
	}

	// Create indexes for efficient lookups.
	_, err = tx.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_org_members_operator ON organization_members(operator_id)
	`)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_org_members_org ON organization_members(organization_id)
	`)

	return err
}

// createInvitationsTable creates the invitations table.
func createInvitationsTable(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS invitations (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL,
			email TEXT NOT NULL,
			role TEXT NOT NULL CHECK (role IN ('admin', 'operator', 'viewer')),
			status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'expired')),
			token TEXT NOT NULL UNIQUE,
			inviter_notes TEXT,
			invitee_notes TEXT,
			invited_by TEXT NOT NULL,
			invited_at INTEGER NOT NULL,
			responded_at INTEGER,
			responder_id TEXT,
			expires_at INTEGER NOT NULL,
			FOREIGN KEY (organization_id) REFERENCES organizations(id),
			FOREIGN KEY (invited_by) REFERENCES operators(id),
			FOREIGN KEY (responder_id) REFERENCES operators(id)
		)
	`)
	if err != nil {
		return err
	}

	// Create indexes for efficient lookups.
	_, err = tx.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_invitations_token ON invitations(token)
	`)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_invitations_email ON invitations(email)
	`)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(context.Background(), `
		CREATE INDEX IF NOT EXISTS idx_invitations_org_status ON invitations(organization_id, status)
	`)

	return err
}

// addOrganizationIDToDevices adds organization_id column to devices table.
func addOrganizationIDToDevices(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
		ALTER TABLE devices ADD COLUMN organization_id TEXT
	`)
	if err != nil {
		// Column may already exist (SQLite ignores duplicate column additions).
		if !isColumnExistsError(err) {
			return err
		}
	}
	return nil
}

// addOrganizationIDToSessions adds organization_id column to sessions table.
func addOrganizationIDToSessions(tx *sql.Tx) error {
	_, err := tx.ExecContext(context.Background(), `
		ALTER TABLE auth_sessions ADD COLUMN organization_id TEXT
	`)
	if err != nil {
		// Column may already exist (SQLite ignores duplicate column additions).
		if !isColumnExistsError(err) {
			return err
		}
	}
	return nil
}

// addOrganizationIDToAPIKeys adds organization_id column to api_keys table.
func addOrganizationIDToAPIKeys(tx *sql.Tx) error {
	// Note: api_keys table might be named differently, check for api_clients.
	// First check if api_keys table exists.
	var tableName string
	err := tx.QueryRowContext(context.Background(), `
		SELECT name FROM sqlite_master WHERE type='table' AND name IN ('api_keys', 'api_clients')
	`).Scan(&tableName)

	if errors.Is(err, sql.ErrNoRows) {
		// Table doesn't exist yet, nothing to migrate.
		return nil
	}
	if err != nil {
		return err
	}

	// Use the actual table name found.
	actualTableName := "api_clients" // default.
	if tableName != "" {
		actualTableName = tableName
	}

	query := `ALTER TABLE ` + actualTableName + ` ADD COLUMN organization_id TEXT`
	_, err = tx.ExecContext(context.Background(), query)
	if err != nil {
		// Column may already exist (SQLite ignores duplicate column additions).
		if !isColumnExistsError(err) {
			return err
		}
	}
	return nil
}
