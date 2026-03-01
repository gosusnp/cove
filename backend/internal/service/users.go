// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/gosusnp/cove/backend/internal/store"
)

// UserService handles user profile operations.
type UserService struct {
	store *store.UserStore
}

// NewUserService returns a new UserService.
func NewUserService(s *store.UserStore) *UserService {
	return &UserService{store: s}
}

// Get returns the user with the given ID.
func (s *UserService) Get(id uuid.UUID) (*store.User, error) {
	user, err := s.store.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

// CreatePAT creates a named PAT for the user. Returns the raw token (shown once) and the PAT metadata.
func (s *UserService) CreatePAT(userID, orgID uuid.UUID, name string) (string, *store.PAT, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", nil, &ValidationError{Msg: "name is required"}
	}
	token, pat, err := s.store.CreatePAT(userID, orgID, name)
	if err != nil {
		return "", nil, fmt.Errorf("create pat: %w", err)
	}
	return token, pat, nil
}

// ListPATs returns all PATs for the user.
func (s *UserService) ListPATs(userID uuid.UUID) ([]store.PAT, error) {
	pats, err := s.store.ListPATs(userID)
	if err != nil {
		return nil, fmt.Errorf("list pats: %w", err)
	}
	return pats, nil
}

// DeletePAT deletes the PAT with the given id for the user.
func (s *UserService) DeletePAT(userID uuid.UUID, id uuid.UUID) error {
	if err := s.store.DeletePAT(userID, id); err != nil {
		return fmt.Errorf("delete pat: %w", err)
	}
	return nil
}
