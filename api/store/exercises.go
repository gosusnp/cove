package store

import (
	"database/sql"
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

type ExerciseStore struct {
	db *sql.DB
}

func NewExerciseStore(db *sql.DB) *ExerciseStore {
	return &ExerciseStore{db: db}
}

func (s *ExerciseStore) List() ([]Exercise, error) {
	rows, err := s.db.Query(`SELECT id, name, progression FROM exercises ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list exercises: %w", err)
	}
	defer rows.Close()

	exercises := []Exercise{}
	for rows.Next() {
		var e Exercise
		if err := rows.Scan(&e.ID, &e.Name, &e.Progression); err != nil {
			return nil, fmt.Errorf("scan exercise: %w", err)
		}
		exercises = append(exercises, e)
	}
	return exercises, rows.Err()
}

func (s *ExerciseStore) Get(id int64) (*Exercise, error) {
	var e Exercise
	err := s.db.QueryRow(`SELECT id, name, progression FROM exercises WHERE id = ?`, id).
		Scan(&e.ID, &e.Name, &e.Progression)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get exercise: %w", err)
	}
	return &e, nil
}

func (s *ExerciseStore) Create(name string, progression *string) (*Exercise, error) {
	res, err := s.db.Exec(
		`INSERT INTO exercises (name, progression) VALUES (?, ?)`,
		name, progression,
	)
	if err != nil {
		return nil, fmt.Errorf("create exercise: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}
	return s.Get(id)
}

func (s *ExerciseStore) Update(id int64, name string, progression *string) (*Exercise, error) {
	res, err := s.db.Exec(
		`UPDATE exercises SET name = ?, progression = ? WHERE id = ?`,
		name, progression, id,
	)
	if err != nil {
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

func (s *ExerciseStore) Delete(id int64) error {
	res, err := s.db.Exec(`DELETE FROM exercises WHERE id = ?`, id)
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
