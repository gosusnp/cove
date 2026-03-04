// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrNotFound = errors.New("not found")
var ErrDuplicate = errors.New("duplicate")

type ExerciseStore struct {
	db *sql.DB
}

func NewExerciseStore(db *sql.DB) *ExerciseStore {
	return &ExerciseStore{db: db}
}

func (s *ExerciseStore) List() ([]domain.ExerciseLite, error) {
	rows, err := s.db.Query(`SELECT id, name FROM exercises ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list exercises: %w", err)
	}
	defer rows.Close()

	exercises := []domain.ExerciseLite{}
	for rows.Next() {
		var e domain.ExerciseLite
		if err := rows.Scan(&e.ID, &e.Name); err != nil {
			return nil, fmt.Errorf("scan exercise: %w", err)
		}
		exercises = append(exercises, e)
	}
	return exercises, rows.Err()
}

func (s *ExerciseStore) Get(id domain.ExerciseID) (*domain.Exercise, error) {
	var e domain.Exercise
	err := s.db.QueryRow(`SELECT id, name, progression FROM exercises WHERE id = $1`, id).
		Scan(&e.ID, &e.Name, &e.Progression)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get exercise: %w", err)
	}
	return &e, nil
}

func isUniqueConstraintErr(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (s *ExerciseStore) Create(name string, progression *string) (*domain.Exercise, error) {
	var id domain.ExerciseID
	err := s.db.QueryRow(
		`INSERT INTO exercises (name, progression) VALUES ($1, $2) RETURNING id`,
		name, progression,
	).Scan(&id)
	if err != nil {
		if isUniqueConstraintErr(err) {
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("create exercise: %w", err)
	}
	return s.Get(id)
}

func (s *ExerciseStore) Update(id domain.ExerciseID, name string, progression *string) (*domain.Exercise, error) {
	res, err := s.db.Exec(
		`UPDATE exercises SET name = $1, progression = $2 WHERE id = $3`,
		name, progression, id,
	)
	if err != nil {
		if isUniqueConstraintErr(err) {
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("update exercise: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return nil, ErrNotFound
	}
	return s.Get(id)
}

func (s *ExerciseStore) Delete(id domain.ExerciseID) error {
	res, err := s.db.Exec(`DELETE FROM exercises WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete exercise: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
