// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import "github.com/gosusnp/cove/backend/internal/store"

type ProgramSetService struct {
	store *store.ProgramSetStore
}

func NewProgramSetService(s *store.ProgramSetStore) *ProgramSetService {
	return &ProgramSetService{store: s}
}

func (s *ProgramSetService) List(programID int64) ([]store.ProgramSet, error) {
	return s.store.List(programID)
}

func (s *ProgramSetService) Get(programID, id int64) (*store.ProgramSet, error) {
	return s.store.Get(programID, id)
}

func (s *ProgramSetService) Create(programID int64, name *string, rounds int, intraSetRestSeconds, sortOrder *int) (*store.ProgramSet, error) {
	if rounds < 1 {
		rounds = 1
	}
	return s.store.Create(programID, name, rounds, intraSetRestSeconds, sortOrder)
}

func (s *ProgramSetService) Update(programID, id int64, name *string, rounds int, intraSetRestSeconds, sortOrder *int) (*store.ProgramSet, error) {
	if rounds < 1 {
		rounds = 1
	}
	return s.store.Update(programID, id, name, rounds, intraSetRestSeconds, sortOrder)
}

func (s *ProgramSetService) Delete(programID, id int64) error {
	return s.store.Delete(programID, id)
}
