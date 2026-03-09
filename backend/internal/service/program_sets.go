// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"context"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/store"
)

type ProgramSetService struct {
	programs *ProgramService
}

func NewProgramSetService(programs *ProgramService) *ProgramSetService {
	return &ProgramSetService{programs: programs}
}

func (s *ProgramSetService) List(ctx context.Context, programID domain.ProgramID) ([]store.ProgramSet, error) {
	return s.programs.ListSets(ctx, programID)
}

func (s *ProgramSetService) Get(ctx context.Context, programID domain.ProgramID, id int64) (*store.ProgramSet, error) {
	return s.programs.GetSet(ctx, programID, id)
}

func (s *ProgramSetService) Create(ctx context.Context, programID domain.ProgramID, name *string, rounds int, intraSetRestSeconds, sortOrder *int) (*store.ProgramSet, error) {
	return s.programs.CreateSet(ctx, programID, name, rounds, intraSetRestSeconds, sortOrder)
}

func (s *ProgramSetService) Update(ctx context.Context, programID domain.ProgramID, id int64, name *string, rounds int, intraSetRestSeconds, sortOrder *int) (*store.ProgramSet, error) {
	return s.programs.UpdateSet(ctx, programID, id, name, rounds, intraSetRestSeconds, sortOrder)
}

func (s *ProgramSetService) Delete(ctx context.Context, programID domain.ProgramID, id int64) error {
	return s.programs.DeleteSet(ctx, programID, id)
}
