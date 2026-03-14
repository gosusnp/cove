// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gosusnp/cove/backend/internal/crypto"
	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/store"
)

type WorkoutSessionService struct {
	db    *sql.DB
	store *store.WorkoutSessionStore
	enc   crypto.Encryptor
}

func NewWorkoutSessionService(db *sql.DB, s *store.WorkoutSessionStore, enc crypto.Encryptor) *WorkoutSessionService {
	return &WorkoutSessionService{db: db, store: s, enc: enc}
}

// attachEncryptor injects the encryptor into a session returned by the store so
// the handler can call UseSensitiveData. The service never decrypts the payload.
func (s *WorkoutSessionService) attachEncryptor(ws *domain.WorkoutSession) {
	ws.SetEncryptor(s.enc)
}

func (s *WorkoutSessionService) encryptSensitiveData(ctx context.Context, p store.WorkoutSessionParams, userID domain.UserID) ([]byte, error) {
	if p.SensitiveData.IsEmpty() {
		return nil, nil
	}
	field := crypto.NewEncryptedField[domain.SessionSensitiveData](s.enc)
	if err := field.Set(ctx, p.SensitiveData, userID.UUID[:]); err != nil {
		return nil, fmt.Errorf("encrypt session sensitive data: %w", err)
	}
	return field.Value(), nil
}

func (s *WorkoutSessionService) List(ctx context.Context) ([]*domain.WorkoutSession, error) {
	id, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	var list []*domain.WorkoutSession
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		list, err = s.store.List(ctx, q, id.OrgID, id.UserID)
		return err
	})
	if err != nil {
		return nil, err
	}
	for _, ws := range list {
		s.attachEncryptor(ws)
	}
	return list, nil
}

func (s *WorkoutSessionService) Get(ctx context.Context, id domain.WorkoutSessionID) (*domain.WorkoutSession, error) {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	var ws *domain.WorkoutSession
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		ws, err = s.store.Get(ctx, q, identity.OrgID, id)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	s.attachEncryptor(ws)
	return ws, nil
}

func (s *WorkoutSessionService) Create(ctx context.Context, p store.WorkoutSessionParams) (*domain.WorkoutSession, error) {
	id, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	sensitiveData, err := s.encryptSensitiveData(ctx, p, id.UserID)
	if err != nil {
		return nil, err
	}

	var ws *domain.WorkoutSession
	err = withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		ws, err = s.store.Create(ctx, q, p, sensitiveData)
		return err
	})
	if err != nil {
		return nil, err
	}
	s.attachEncryptor(ws)
	return ws, nil
}

func (s *WorkoutSessionService) Update(ctx context.Context, id domain.WorkoutSessionID, p store.WorkoutSessionParams) (*domain.WorkoutSession, error) {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	sensitiveData, err := s.encryptSensitiveData(ctx, p, identity.UserID)
	if err != nil {
		return nil, err
	}

	var ws *domain.WorkoutSession
	err = withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		ws, err = s.store.Update(ctx, q, identity.OrgID, id, p, sensitiveData)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	s.attachEncryptor(ws)
	return ws, nil
}

func (s *WorkoutSessionService) Delete(ctx context.Context, id domain.WorkoutSessionID) error {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return ErrUnauthorized
	}

	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		return s.store.Delete(ctx, q, identity.OrgID, id)
	})
	if errors.Is(err, store.ErrNotFound) {
		return ErrNotFound
	}
	return err
}
