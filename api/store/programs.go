package store

import (
	"database/sql"
	"errors"
	"fmt"
)

type ProgramStore struct {
	db *sql.DB
}

func NewProgramStore(db *sql.DB) *ProgramStore {
	return &ProgramStore{db: db}
}

func (s *ProgramStore) List() ([]Program, error) {
	rows, err := s.db.Query(`SELECT id, name FROM programs ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list programs: %w", err)
	}
	defer rows.Close()

	programs := []Program{}
	for rows.Next() {
		var p Program
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			return nil, fmt.Errorf("scan program: %w", err)
		}
		programs = append(programs, p)
	}
	return programs, rows.Err()
}

func (s *ProgramStore) Get(id int64) (*Program, error) {
	var p Program
	err := s.db.QueryRow(`SELECT id, name FROM programs WHERE id = ?`, id).
		Scan(&p.ID, &p.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get program: %w", err)
	}
	return &p, nil
}

func (s *ProgramStore) Create(name string) (*Program, error) {
	res, err := s.db.Exec(`INSERT INTO programs (name) VALUES (?)`, name)
	if err != nil {
		return nil, fmt.Errorf("create program: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}
	return s.Get(id)
}

func (s *ProgramStore) Update(id int64, name string) (*Program, error) {
	res, err := s.db.Exec(`UPDATE programs SET name = ? WHERE id = ?`, name, id)
	if err != nil {
		return nil, fmt.Errorf("update program: %w", err)
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

func (s *ProgramStore) Delete(id int64) error {
	res, err := s.db.Exec(`DELETE FROM programs WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete program: %w", err)
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
