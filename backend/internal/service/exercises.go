// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"errors"
	"strings"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/store"
)

func normalizeName(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

type ExerciseService struct {
	store *store.ExerciseStore
}

func NewExerciseService(s *store.ExerciseStore) *ExerciseService {
	return &ExerciseService{store: s}
}

func (s *ExerciseService) List() ([]domain.ExerciseLite, error) {
	return s.store.List()
}

func (s *ExerciseService) Get(id domain.ExerciseID) (*domain.Exercise, error) {
	e, err := s.store.Get(id)
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return e, err
}

func (s *ExerciseService) Create(name string, progression *string) (*domain.Exercise, error) {
	name = normalizeName(name)
	if name == "" {
		return nil, &ValidationError{Msg: "name is required"}
	}
	e, err := s.store.Create(name, progression)
	if errors.Is(err, store.ErrDuplicate) {
		return nil, &ValidationError{Msg: "exercise with this name already exists"}
	}
	return e, err
}

func (s *ExerciseService) Update(id domain.ExerciseID, name string, progression *string) (*domain.Exercise, error) {
	name = normalizeName(name)
	if name == "" {
		return nil, &ValidationError{Msg: "name is required"}
	}
	e, err := s.store.Update(id, name, progression)
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	if errors.Is(err, store.ErrDuplicate) {
		return nil, &ValidationError{Msg: "exercise with this name already exists"}
	}
	return e, err
}

func (s *ExerciseService) Delete(id domain.ExerciseID) error {
	if err := s.store.Delete(id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}
