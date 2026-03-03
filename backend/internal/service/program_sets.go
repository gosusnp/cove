// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"errors"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/store"
)

type ProgramSetService struct {
	store *store.ProgramSetStore
}

func NewProgramSetService(s *store.ProgramSetStore) *ProgramSetService {
	return &ProgramSetService{store: s}
}

func (s *ProgramSetService) List(programID domain.ProgramID) ([]store.ProgramSet, error) {
	return s.store.List(programID)
}

func (s *ProgramSetService) Get(programID domain.ProgramID, id int64) (*store.ProgramSet, error) {
	ps, err := s.store.Get(programID, id)
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return ps, err
}

func (s *ProgramSetService) Create(programID domain.ProgramID, name *string, rounds int, intraSetRestSeconds, sortOrder *int) (*store.ProgramSet, error) {
	if rounds < 1 {
		rounds = 1
	}
	return s.store.Create(programID, name, rounds, intraSetRestSeconds, sortOrder)
}

func (s *ProgramSetService) Update(programID domain.ProgramID, id int64, name *string, rounds int, intraSetRestSeconds, sortOrder *int) (*store.ProgramSet, error) {
	if rounds < 1 {
		rounds = 1
	}
	ps, err := s.store.Update(programID, id, name, rounds, intraSetRestSeconds, sortOrder)
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return ps, err
}

func (s *ProgramSetService) Delete(programID domain.ProgramID, id int64) error {
	if err := s.store.Delete(programID, id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}
