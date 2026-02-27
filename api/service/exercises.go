// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"errors"
	"strings"

	"github.com/gosusnp/cove/api/store"
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

func (s *ExerciseService) List() ([]store.Exercise, error) {
	return s.store.List()
}

func (s *ExerciseService) Get(id int64) (*store.ExerciseDetail, error) {
	return s.store.Get(id)
}

func (s *ExerciseService) Create(name string, progression *string) (*store.ExerciseDetail, error) {
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

func (s *ExerciseService) Update(id int64, name string, progression *string) (*store.ExerciseDetail, error) {
	name = normalizeName(name)
	if name == "" {
		return nil, &ValidationError{Msg: "name is required"}
	}
	e, err := s.store.Update(id, name, progression)
	if errors.Is(err, store.ErrDuplicate) {
		return nil, &ValidationError{Msg: "exercise with this name already exists"}
	}
	return e, err
}

func (s *ExerciseService) Delete(id int64) error {
	return s.store.Delete(id)
}
