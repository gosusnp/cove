// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"fmt"

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
