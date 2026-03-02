// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/gosusnp/cove/backend/internal/store"
)

// UserService handles user profile operations.
type UserService struct {
	db    *sql.DB
	users *store.UserStore
	orgs  *store.OrgStore
}

// NewUserService returns a new UserService.
func NewUserService(db *sql.DB, users *store.UserStore, orgs *store.OrgStore) *UserService {
	return &UserService{db: db, users: users, orgs: orgs}
}

// GetOrCreate upserts a user by google_sub, creating an org and membership on first insert.
// Returns the user and whether it was newly created.
func (s *UserService) GetOrCreate(email, googleSub string) (*store.User, bool, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, false, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	newID, err := uuid.NewV7()
	if err != nil {
		return nil, false, fmt.Errorf("generate user id: %w", err)
	}

	txUsers := s.users.WithTx(tx)
	txOrgs := s.orgs.WithTx(tx)

	user, created, err := txUsers.UpsertUser(newID, email, googleSub)
	if err != nil {
		return nil, false, err
	}

	if created {
		orgID, err := uuid.NewV7()
		if err != nil {
			return nil, false, fmt.Errorf("generate org id: %w", err)
		}
		if err := txOrgs.CreateOrg(orgID, email); err != nil {
			return nil, false, err
		}
		if err := txOrgs.CreateOrgMember(orgID, user.ID, "owner"); err != nil {
			return nil, false, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, false, fmt.Errorf("commit: %w", err)
	}
	return user, created, nil
}

// CreateSession generates a session token for the user.
func (s *UserService) CreateSession(userID uuid.UUID, ip, browser, os string) (string, error) {
	token, err := s.users.CreateSession(userID, ip, browser, os)
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	return token, nil
}

// Get returns the user with the given ID.
func (s *UserService) Get(id uuid.UUID) (*store.User, error) {
	user, err := s.users.GetByID(id)
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

// CreatePAT creates a named PAT for the user. Returns the raw token (shown once) and the PAT metadata.
func (s *UserService) CreatePAT(userID, orgID uuid.UUID, name, ipMasked, browser, os string) (string, *store.PAT, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", nil, &ValidationError{Msg: "name is required"}
	}
	token, pat, err := s.users.CreatePAT(userID, orgID, name, ipMasked, browser, os)
	if err != nil {
		return "", nil, fmt.Errorf("create pat: %w", err)
	}
	return token, pat, nil
}

// ListPATs returns all PATs for the user.
func (s *UserService) ListPATs(userID uuid.UUID) ([]store.PAT, error) {
	pats, err := s.users.ListPATs(userID)
	if err != nil {
		return nil, fmt.Errorf("list pats: %w", err)
	}
	return pats, nil
}

// ListSessions returns all active sessions for the user.
func (s *UserService) ListSessions(userID uuid.UUID) ([]store.Session, error) {
	sessions, err := s.users.ListSessions(userID)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	return sessions, nil
}

// DeletePAT deletes the PAT with the given id for the user.
func (s *UserService) DeletePAT(userID uuid.UUID, id uuid.UUID) error {
	if err := s.users.DeletePAT(userID, id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("delete pat: %w", err)
	}
	return nil
}

// DeleteSession deletes the session with the given id for the user.
func (s *UserService) DeleteSession(userID uuid.UUID, id uuid.UUID) error {
	if err := s.users.DeleteSession(userID, id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
