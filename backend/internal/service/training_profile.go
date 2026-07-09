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

type TrainingProfileService struct {
	db    *sql.DB
	store *store.TrainingProfileStore
	enc   crypto.Encryptor
}

func NewTrainingProfileService(db *sql.DB, s *store.TrainingProfileStore, enc crypto.Encryptor) *TrainingProfileService {
	return &TrainingProfileService{db: db, store: s, enc: enc}
}

// TrainingProfilePatch contains fields for partial updates to a training profile.
type TrainingProfilePatch struct {
	Motivation  domain.Optional[*string]                          `json:"motivation"`
	Disciplines domain.Optional[[]TrainingProfileDisciplinePatch] `json:"disciplines"`
	Constraints domain.Optional[*string]                          `json:"constraints"`
}

type TrainingProfileDisciplinePatch struct {
	Name          *string  `json:"name,omitempty"`
	YearsPractice *float64 `json:"years_practice,omitempty"`
	Level         *string  `json:"level,omitempty"`
	Notes         *string  `json:"notes,omitempty"`
}

func (s *TrainingProfileService) attachEncryptor(tp *domain.UserTrainingProfile) {
	tp.SetEncryptor(s.enc)
}

func (s *TrainingProfileService) encryptSensitiveData(ctx context.Context, data domain.TrainingProfileSensitiveData, userID domain.UserID) ([]byte, error) {
	if data.IsEmpty() {
		return nil, nil
	}
	field := crypto.NewEncryptedField[domain.TrainingProfileSensitiveData](s.enc)
	if err := field.Set(ctx, data, userID.UUID[:]); err != nil {
		return nil, fmt.Errorf("encrypt training profile: %w", err)
	}
	return field.Value(), nil
}

func (s *TrainingProfileService) Get(ctx context.Context) (*domain.UserTrainingProfile, error) {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	var tp *domain.UserTrainingProfile
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		tp, err = s.store.Get(ctx, q, identity.UserID)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	s.attachEncryptor(tp)
	return tp, nil
}

func (s *TrainingProfileService) Upsert(ctx context.Context, data domain.TrainingProfileSensitiveData) (*domain.UserTrainingProfile, error) {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	sensitiveData, err := s.encryptSensitiveData(ctx, data, identity.UserID)
	if err != nil {
		return nil, err
	}

	var tp *domain.UserTrainingProfile
	err = withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		tp, err = s.store.Upsert(ctx, q, identity.UserID, identity.OrgID, sensitiveData)
		return err
	})
	if err != nil {
		return nil, err
	}
	s.attachEncryptor(tp)
	return tp, nil
}

func (s *TrainingProfileService) Patch(ctx context.Context, patch TrainingProfilePatch) (*domain.UserTrainingProfile, error) {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	var tp *domain.UserTrainingProfile
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		current, err := s.store.Get(ctx, q, identity.UserID)
		if err != nil && !errors.Is(err, store.ErrNotFound) {
			return err
		}

		var data domain.TrainingProfileSensitiveData
		if err == nil {
			s.attachEncryptor(current)
			if err := current.UseSensitiveData(ctx, func(cur domain.TrainingProfileSensitiveData) error {
				data.Motivation = cloneSensitiveString(cur.Motivation)
				data.Constraints = cloneSensitiveString(cur.Constraints)
				data.Disciplines = make([]domain.TrainingProfileDiscipline, len(cur.Disciplines))
				for i, d := range cur.Disciplines {
					data.Disciplines[i] = domain.TrainingProfileDiscipline{
						Name:          cloneSensitiveString(d.Name),
						YearsPractice: d.YearsPractice,
						Level:         cloneSensitiveString(d.Level),
						Notes:         cloneSensitiveString(d.Notes),
					}
				}
				return nil
			}); err != nil {
				return fmt.Errorf("merge sensitive data: %w", err)
			}
		}

		if patch.Motivation.Set {
			data.Motivation = crypto.NewSensitiveStringFromPtr(patch.Motivation.Value)
		}
		if patch.Constraints.Set {
			data.Constraints = crypto.NewSensitiveStringFromPtr(patch.Constraints.Value)
		}
		if patch.Disciplines.Set {
			data.Disciplines = make([]domain.TrainingProfileDiscipline, len(patch.Disciplines.Value))
			for i, dp := range patch.Disciplines.Value {
				data.Disciplines[i] = domain.TrainingProfileDiscipline{
					Name:          crypto.NewSensitiveStringFromPtr(dp.Name),
					YearsPractice: dp.YearsPractice,
					Level:         crypto.NewSensitiveStringFromPtr(dp.Level),
					Notes:         crypto.NewSensitiveStringFromPtr(dp.Notes),
				}
			}
		}

		sensitiveData, err := s.encryptSensitiveData(ctx, data, identity.UserID)
		if err != nil {
			return err
		}

		tp, err = s.store.Upsert(ctx, q, identity.UserID, identity.OrgID, sensitiveData)
		return err
	})
	if err != nil {
		return nil, err
	}
	s.attachEncryptor(tp)
	return tp, nil
}

func (s *TrainingProfileService) Delete(ctx context.Context) error {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return ErrUnauthorized
	}

	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		return s.store.Delete(ctx, q, identity.UserID)
	})
	if errors.Is(err, store.ErrNotFound) {
		return ErrNotFound
	}
	return err
}
