// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"context"
	"fmt"

	"github.com/gosusnp/cove/backend/internal/domain"
)

type OrgStore struct{}

func NewOrgStore() *OrgStore {
	return &OrgStore{}
}

// CreateOrg inserts a new org with the given id and name.
func (s *OrgStore) CreateOrg(
	ctx context.Context,
	q Querier,
	id domain.OrgID,
	name string,
) error {
	if _, err := q.ExecContext(ctx, `INSERT INTO orgs (id, name) VALUES ($1, $2)`, id, name); err != nil {
		return fmt.Errorf("create org: %w", err)
	}
	return nil
}

// CreateOrgMember inserts a membership linking orgID and userID with the given role.
func (s *OrgStore) CreateOrgMember(
	ctx context.Context,
	q Querier,
	orgID domain.OrgID,
	userID domain.UserID,
	role string,
) error {
	if _, err := q.Exec(`INSERT INTO org_members (org_id, user_id, role) VALUES ($1, $2, $3)`, orgID, userID, role); err != nil {
		return fmt.Errorf("create org member: %w", err)
	}
	return nil
}
