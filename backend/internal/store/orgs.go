// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type OrgStore struct {
	baseStore
}

func NewOrgStore(db *sql.DB) *OrgStore {
	return &OrgStore{baseStore{db: db}}
}

// WithTx returns an OrgStore that executes queries within tx.
func (s *OrgStore) WithTx(tx *sql.Tx) *OrgStore {
	return &OrgStore{s.withTx(tx)}
}

// CreateOrg inserts a new org with the given id and name.
func (s *OrgStore) CreateOrg(id uuid.UUID, name string) error {
	if _, err := s.db.Exec(`INSERT INTO orgs (id, name) VALUES ($1, $2)`, id, name); err != nil {
		return fmt.Errorf("create org: %w", err)
	}
	return nil
}

// CreateOrgMember inserts a membership linking orgID and userID with the given role.
func (s *OrgStore) CreateOrgMember(orgID, userID uuid.UUID, role string) error {
	if _, err := s.db.Exec(`INSERT INTO org_members (org_id, user_id, role) VALUES ($1, $2, $3)`, orgID, userID, role); err != nil {
		return fmt.Errorf("create org member: %w", err)
	}
	return nil
}
