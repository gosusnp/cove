// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gosusnp/cove/backend/internal/domain"
)

type ProgramSetStore struct {
	db *sql.DB
}

func NewProgramSetStore(db *sql.DB) *ProgramSetStore {
	return &ProgramSetStore{db: db}
}

const programSetColumns = `id, program_id, name, rounds, intra_set_rest_seconds, sort_order`

func scanProgramSet(row interface{ Scan(...any) error }) (*ProgramSet, error) {
	var ps ProgramSet
	if err := row.Scan(&ps.ID, &ps.ProgramID, &ps.Name, &ps.Rounds, &ps.IntraSetRestSeconds, &ps.SortOrder); err != nil {
		return nil, err
	}
	return &ps, nil
}

func (s *ProgramSetStore) List(programID domain.ProgramID) ([]ProgramSet, error) {
	rows, err := s.db.Query(
		`SELECT `+programSetColumns+` FROM program_sets WHERE program_id = $1 ORDER BY sort_order, id`,
		programID,
	)
	if err != nil {
		return nil, fmt.Errorf("list program sets: %w", err)
	}
	defer rows.Close()

	sets := []ProgramSet{}
	for rows.Next() {
		ps, err := scanProgramSet(rows)
		if err != nil {
			return nil, fmt.Errorf("scan program set: %w", err)
		}
		sets = append(sets, *ps)
	}
	return sets, rows.Err()
}

func (s *ProgramSetStore) Get(programID domain.ProgramID, id int64) (*ProgramSet, error) {
	row := s.db.QueryRow(
		`SELECT `+programSetColumns+` FROM program_sets WHERE id = $1 AND program_id = $2`,
		id, programID,
	)
	ps, err := scanProgramSet(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get program set: %w", err)
	}
	return ps, nil
}

func (s *ProgramSetStore) Create(programID domain.ProgramID, name *string, rounds int, intraSetRestSeconds, sortOrder *int) (*ProgramSet, error) {
	var id int64
	err := s.db.QueryRow(
		`INSERT INTO program_sets (program_id, name, rounds, intra_set_rest_seconds, sort_order) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		programID, name, rounds, intraSetRestSeconds, sortOrder,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create program set: %w", err)
	}
	return s.Get(programID, id)
}

func (s *ProgramSetStore) Update(programID domain.ProgramID, id int64, name *string, rounds int, intraSetRestSeconds, sortOrder *int) (*ProgramSet, error) {
	res, err := s.db.Exec(
		`UPDATE program_sets SET name = $1, rounds = $2, intra_set_rest_seconds = $3, sort_order = $4 WHERE id = $5 AND program_id = $6`,
		name, rounds, intraSetRestSeconds, sortOrder, id, programID,
	)
	if err != nil {
		return nil, fmt.Errorf("update program set: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return nil, ErrNotFound
	}
	return s.Get(programID, id)
}

func (s *ProgramSetStore) Delete(programID domain.ProgramID, id int64) error {
	res, err := s.db.Exec(
		`DELETE FROM program_sets WHERE id = $1 AND program_id = $2`,
		id, programID,
	)
	if err != nil {
		return fmt.Errorf("delete program set: %w", err)
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
