package service

import "github.com/gosusnp/cove/api/store"

type ProgramService struct {
	store *store.ProgramStore
}

func NewProgramService(s *store.ProgramStore) *ProgramService {
	return &ProgramService{store: s}
}

func (s *ProgramService) List() ([]store.Program, error) {
	return s.store.List()
}

func (s *ProgramService) GetDetail(id int64) (*store.ProgramDetail, error) {
	return s.store.GetDetail(id)
}

func (s *ProgramService) Create(name string) (*store.Program, error) {
	if name == "" {
		return nil, &ValidationError{Msg: "name is required"}
	}
	return s.store.Create(name)
}

func (s *ProgramService) Update(id int64, name string) (*store.Program, error) {
	if name == "" {
		return nil, &ValidationError{Msg: "name is required"}
	}
	return s.store.Update(id, name)
}

func (s *ProgramService) Delete(id int64) error {
	return s.store.Delete(id)
}
