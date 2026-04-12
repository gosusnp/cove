// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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

// WorkoutSessionPatch contains the fields that may be changed via a partial
// update. Only fields where Optional.Set == true are applied; absent fields
// retain their current values unchanged. UpdatedAt is an optional version token
// for optimistic locking; when provided, the update is rejected if the stored
// updated_at does not match. When nil, the update is applied unconditionally
// (last-write-wins).
type WorkoutSessionPatch struct {
	UpdatedAt        *time.Time                  `json:"updated_at,omitempty"`
	ProgramID        domain.Optional[*int64]     `json:"program_id"`
	Activity         domain.Optional[*string]    `json:"activity"`
	DurationS        domain.Optional[*int]       `json:"duration_s"`
	StartedAt        domain.Optional[*time.Time] `json:"started_at"`
	CompletedAt      domain.Optional[*time.Time] `json:"completed_at"`
	PerceivedEffort  domain.Optional[*int]       `json:"perceived_effort"`
	SessionNotes     domain.Optional[*string]    `json:"session_notes"`
	ProgramName      domain.Optional[*string]    `json:"program_name"`
	ProgramStructure domain.Optional[*string]    `json:"program_structure"`
	Summary          domain.Optional[*string]    `json:"summary"`
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

// cloneSensitiveString returns a copy of s, or nil if s is nil.
func cloneSensitiveString(s *crypto.SensitiveString) *crypto.SensitiveString {
	if s == nil {
		return nil
	}
	ss := crypto.NewSensitiveString(s.String())
	return &ss
}

// newSensitiveStringPtr wraps s as a SensitiveString, or returns nil if s is nil.
func newSensitiveStringPtr(s *string) *crypto.SensitiveString {
	if s == nil {
		return nil
	}
	ss := crypto.NewSensitiveString(*s)
	return &ss
}

// SessionFilter holds optional date range bounds for listing sessions.
// From is an inclusive lower bound; To is an exclusive upper bound on started_at.
// Sessions with a null started_at are excluded when either bound is set.
type SessionFilter struct {
	From *time.Time
	To   *time.Time
}

// NewSessionFilter parses optional YYYY-MM-DD date strings into a SessionFilter.
// from is an inclusive lower bound; to is converted to an exclusive next-day upper bound.
func NewSessionFilter(from, to *string) (SessionFilter, error) {
	var f SessionFilter
	if from != nil {
		t, err := time.ParseInLocation("2006-01-02", *from, time.UTC)
		if err != nil {
			return f, errors.New("invalid from date: use YYYY-MM-DD")
		}
		f.From = &t
	}
	if to != nil {
		t, err := time.ParseInLocation("2006-01-02", *to, time.UTC)
		if err != nil {
			return f, errors.New("invalid to date: use YYYY-MM-DD")
		}
		t = t.AddDate(0, 0, 1)
		f.To = &t
	}
	return f, nil
}

func (s *WorkoutSessionService) List(ctx context.Context, f SessionFilter) ([]*domain.WorkoutSession, error) {
	id, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	var list []*domain.WorkoutSession
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		list, err = s.store.List(ctx, q, id.OrgID, id.UserID, f.From, f.To)
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

func (s *WorkoutSessionService) Update(ctx context.Context, id domain.WorkoutSessionID, updatedAt *time.Time, p store.WorkoutSessionParams) (*domain.WorkoutSession, error) {
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
		ws, err = s.store.Update(ctx, q, identity.OrgID, id, p, sensitiveData, p.SensitiveData.Summary != nil, updatedAt)
		return err
	})
	if errors.Is(err, store.ErrConflict) {
		return nil, ErrConflict
	}
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	s.attachEncryptor(ws)
	return ws, nil
}

// Patch applies a partial update to a workout session. Only fields where
// Optional.Set == true are changed; all others retain their current values.
// The read-modify-write for sensitive data is handled entirely within this
// method — callers pass a patch spec and do not interact with the encrypted
// blob directly.
func (s *WorkoutSessionService) Patch(ctx context.Context, id domain.WorkoutSessionID, patch WorkoutSessionPatch) (*domain.WorkoutSession, error) {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	var ws *domain.WorkoutSession
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		current, err := s.store.Get(ctx, q, identity.OrgID, id)
		if err != nil {
			return err
		}
		s.attachEncryptor(current)

		// Start from current non-sensitive values.
		p := store.WorkoutSessionParams{
			ProgramID:   current.ProgramID,
			Activity:    current.Activity,
			DurationS:   current.DurationS,
			StartedAt:   current.StartedAt,
			CompletedAt: current.CompletedAt,
		}

		// Apply non-sensitive patches.
		if patch.ProgramID.Set {
			if patch.ProgramID.Value != nil {
				pid := domain.ProgramID(*patch.ProgramID.Value)
				p.ProgramID = &pid
			} else {
				p.ProgramID = nil
			}
		}
		if patch.Activity.Set {
			p.Activity = patch.Activity.Value
		}
		if patch.DurationS.Set {
			p.DurationS = patch.DurationS.Value
		}
		if patch.StartedAt.Set {
			p.StartedAt = patch.StartedAt.Value
		}
		if patch.CompletedAt.Set {
			p.CompletedAt = patch.CompletedAt.Value
		}

		// Merge sensitive fields: clone current values, then apply patches.
		setSummaryNow := false
		if err := current.UseSensitiveData(ctx, func(cur domain.SessionSensitiveData) error {
			p.SensitiveData.PerceivedEffort = cur.PerceivedEffort
			p.SensitiveData.SessionNotes = cloneSensitiveString(cur.SessionNotes)
			p.SensitiveData.ProgramName = cloneSensitiveString(cur.ProgramName)
			p.SensitiveData.ProgramStructure = cloneSensitiveString(cur.ProgramStructure)
			p.SensitiveData.Summary = cloneSensitiveString(cur.Summary)

			if patch.PerceivedEffort.Set {
				p.SensitiveData.PerceivedEffort = patch.PerceivedEffort.Value
			}
			if patch.SessionNotes.Set {
				p.SensitiveData.SessionNotes = newSensitiveStringPtr(patch.SessionNotes.Value)
			}
			if patch.ProgramName.Set {
				p.SensitiveData.ProgramName = newSensitiveStringPtr(patch.ProgramName.Value)
			}
			if patch.ProgramStructure.Set {
				p.SensitiveData.ProgramStructure = newSensitiveStringPtr(patch.ProgramStructure.Value)
			}
			if patch.Summary.Set {
				p.SensitiveData.Summary = newSensitiveStringPtr(patch.Summary.Value)
				setSummaryNow = patch.Summary.Value != nil
			}
			return nil
		}); err != nil {
			return fmt.Errorf("merge sensitive data: %w", err)
		}

		sensitiveData, err := s.encryptSensitiveData(ctx, p, identity.UserID)
		if err != nil {
			return err
		}

		ws, err = s.store.Update(ctx, q, identity.OrgID, id, p, sensitiveData, setSummaryNow, patch.UpdatedAt)
		return err
	})
	if errors.Is(err, store.ErrConflict) {
		return nil, ErrConflict
	}
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
