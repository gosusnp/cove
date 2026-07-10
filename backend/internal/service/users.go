// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/store"
)

// SessionTTL is the lifetime of a user session. The store uses it as expires_at;
// the handler uses it to set the cookie MaxAge.
const SessionTTL = 30 * 24 * time.Hour

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
func (s *UserService) GetOrCreate(ctx context.Context, email domain.Email, googleSub domain.GoogleSub) (*domain.User, bool, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, false, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	newID := domain.NewUserID()

	user, created, err := s.users.UpsertUser(ctx, tx, newID, email, googleSub)
	if err != nil {
		return nil, false, err
	}

	if created {
		orgID := domain.NewOrgID()
		if err := s.orgs.CreateOrg(ctx, tx, orgID, string(email)); err != nil {
			return nil, false, err
		}
		// TODO remove userID wrapping
		if err := s.orgs.CreateOrgMember(ctx, tx, orgID, domain.UserID(user.ID), "owner"); err != nil {
			return nil, false, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, false, fmt.Errorf("commit: %w", err)
	}
	return user, created, nil
}

// Get returns the user with the given ID.
func (s *UserService) Get(ctx context.Context, id domain.UserID) (*domain.User, error) {
	user, err := s.users.GetByID(ctx, s.db, id)
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func (s *UserService) GetUserByToken(
	ctx context.Context,
	token string,
	ipMasked domain.MaskedIP,
	browser string,
	os string,
) (*domain.User, *domain.Org, uuid.UUID, error) {
	return s.users.GetUserByToken(ctx, s.db, token, ipMasked, browser, os)
}

// CreateSession generates a session token for the user and returns the raw token,
// session ID, expiry time, and any error.
func (s *UserService) CreateSession(
	ctx context.Context,
	userID domain.UserID,
	ip domain.MaskedIP,
	browser string,
	os string,
) (string, domain.SessionID, time.Time, error) {
	expiresAt := time.Now().Add(SessionTTL)
	token, sessionID, err := s.users.CreateSession(ctx, s.db, userID, expiresAt, ip, browser, os)
	if err != nil {
		return "", domain.SessionID{}, time.Time{}, fmt.Errorf("create session: %w", err)
	}
	return token, sessionID, expiresAt, nil
}

// CreatePAT creates a named PAT for the user. Returns the raw token (shown once) and the PAT metadata.
func (s *UserService) CreatePAT(
	ctx context.Context,
	userID domain.UserID,
	orgID domain.OrgID,
	name string,
	ipMasked domain.MaskedIP,
	browser string,
	os string,
) (string, *domain.PAT, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", nil, &ValidationError{Msg: "name is required"}
	}
	token, pat, err := s.users.CreatePAT(ctx, s.db, userID, orgID, name, ipMasked, browser, os)
	if err != nil {
		return "", nil, fmt.Errorf("create pat: %w", err)
	}
	return token, pat, nil
}

// ListPATs returns all PATs for the user.
func (s *UserService) ListPATs(ctx context.Context, userID domain.UserID) ([]domain.PAT, error) {
	pats, err := s.users.ListPATs(ctx, s.db, userID)
	if err != nil {
		return nil, fmt.Errorf("list pats: %w", err)
	}
	return pats, nil
}

// ListSessions returns all active sessions for the user.
func (s *UserService) ListSessions(ctx context.Context, userID domain.UserID) ([]domain.Session, error) {
	sessions, err := s.users.ListSessions(ctx, s.db, userID)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	return sessions, nil
}

// DeletePAT deletes the PAT with the given id for the user.
func (s *UserService) DeletePAT(ctx context.Context, userID domain.UserID, id domain.TokenID) error {
	if err := s.users.DeletePAT(ctx, s.db, userID, id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("delete pat: %w", err)
	}
	return nil
}

// UserPreferencesPatch contains the fields that may be changed via a partial
// update. Only fields where Optional.Set == true are applied; absent fields
// retain their current values.
type UserPreferencesPatch struct {
	FitnessUnitSystem domain.Optional[*domain.UnitSystem] `json:"fitness_unit_system"`
	CookingUnitSystem domain.Optional[*domain.UnitSystem] `json:"cooking_unit_system"`
	DisplayName       domain.Optional[*string]            `json:"display_name"`
	FirstName         domain.Optional[*string]            `json:"first_name"`
	LastName          domain.Optional[*string]            `json:"last_name"`
}

// PatchPreferences applies a partial update to the user's unit system preferences.
// Only fields where Optional.Set == true are changed; all others retain their current values.
func (s *UserService) PatchPreferences(ctx context.Context, userID domain.UserID, patch UserPreferencesPatch) (*domain.User, error) {
	current, err := s.users.GetByID(ctx, s.db, userID)
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	newFitness := current.FitnessUnitSystem
	newCooking := current.CookingUnitSystem
	newDisplayName := current.DisplayName
	newFirstName := current.FirstName
	newLastName := current.LastName

	if patch.FitnessUnitSystem.Set {
		newFitness = patch.FitnessUnitSystem.Value
	}
	if patch.CookingUnitSystem.Set {
		newCooking = patch.CookingUnitSystem.Value
	}
	if patch.DisplayName.Set {
		newDisplayName = trimToNil(patch.DisplayName.Value)
	}
	if patch.FirstName.Set {
		newFirstName = trimToNil(patch.FirstName.Value)
	}
	if patch.LastName.Set {
		newLastName = trimToNil(patch.LastName.Value)
	}

	if newFitness != nil {
		if !newFitness.Valid() || *newFitness == domain.UnitSystemUSCustomary {
			return nil, &ValidationError{Msg: "fitness_unit_system must be 'metric' or 'imperial'"}
		}
	}
	if newCooking != nil && !newCooking.Valid() {
		return nil, &ValidationError{Msg: "cooking_unit_system must be 'metric', 'imperial', or 'us_customary'"}
	}

	user, err := s.users.UpdatePreferences(ctx, s.db, userID, newFitness, newCooking, newDisplayName, newFirstName, newLastName)
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update preferences: %w", err)
	}
	return user, nil
}

// trimToNil trims whitespace from s and returns nil if the result is empty.
func trimToNil(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}

// CreateServiceAccount creates a named service account user.
func (s *UserService) CreateServiceAccount(ctx context.Context, name string) (*domain.User, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, &ValidationError{Msg: "name is required"}
	}
	id := domain.NewUserID()
	user, err := s.users.CreateServiceAccount(ctx, s.db, id, name)
	if err != nil {
		return nil, fmt.Errorf("create service account: %w", err)
	}
	return user, nil
}

// ListServiceAccounts returns all service account users.
func (s *UserService) ListServiceAccounts(ctx context.Context) ([]domain.User, error) {
	accounts, err := s.users.ListServiceAccounts(ctx, s.db)
	if err != nil {
		return nil, fmt.Errorf("list service accounts: %w", err)
	}
	return accounts, nil
}

// DeleteServiceAccount deletes a service account and all its tokens.
func (s *UserService) DeleteServiceAccount(ctx context.Context, id domain.UserID) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.users.DeleteServiceAccountPATs(ctx, tx, id); err != nil {
		return fmt.Errorf("delete service account pats: %w", err)
	}
	if err := s.users.DeleteServiceAccount(ctx, tx, id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("delete service account: %w", err)
	}
	return tx.Commit()
}

// requireServiceAccount returns ErrNotFound if the user does not exist or is not a service account.
func (s *UserService) requireServiceAccount(ctx context.Context, id domain.UserID) error {
	u, err := s.users.GetByID(ctx, s.db, id)
	if errors.Is(err, store.ErrNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if !u.IsServiceAccount {
		return ErrNotFound
	}
	return nil
}

// CreateServiceAccountPAT generates a PAT for a service account. Returns the raw token (shown once) and metadata.
func (s *UserService) CreateServiceAccountPAT(ctx context.Context, userID domain.UserID, name string) (string, *domain.PAT, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", nil, &ValidationError{Msg: "name is required"}
	}
	if err := s.requireServiceAccount(ctx, userID); err != nil {
		return "", nil, err
	}
	token, pat, err := s.users.CreateServiceAccountPAT(ctx, s.db, userID, name)
	if err != nil {
		return "", nil, fmt.Errorf("create service account pat: %w", err)
	}
	return token, pat, nil
}

// ListServiceAccountPATs returns all PATs for a service account.
func (s *UserService) ListServiceAccountPATs(ctx context.Context, userID domain.UserID) ([]domain.PAT, error) {
	if err := s.requireServiceAccount(ctx, userID); err != nil {
		return nil, err
	}
	pats, err := s.users.ListPATs(ctx, s.db, userID)
	if err != nil {
		return nil, fmt.Errorf("list service account pats: %w", err)
	}
	return pats, nil
}

// DeleteServiceAccountPAT deletes a PAT belonging to a service account.
func (s *UserService) DeleteServiceAccountPAT(ctx context.Context, userID domain.UserID, tokenID domain.TokenID) error {
	if err := s.requireServiceAccount(ctx, userID); err != nil {
		return err
	}
	if err := s.users.DeletePAT(ctx, s.db, userID, tokenID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("delete service account pat: %w", err)
	}
	return nil
}

// DeleteSession deletes the session with the given id for the user.
func (s *UserService) DeleteSession(ctx context.Context, userID domain.UserID, sessionID domain.SessionID) error {
	if err := s.users.DeleteSession(ctx, s.db, userID, sessionID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
