package service

import (
	"errors"
	"testing"

	"github.com/gosusnp/cove/api/store"
)

func newTestProgramService(t *testing.T) *ProgramService {
	t.Helper()
	return NewProgramService(store.NewProgramStore(newTestDB(t)))
}

func TestProgramService_Create(t *testing.T) {
	t.Run("empty name returns ValidationError", func(t *testing.T) {
		svc := newTestProgramService(t)

		_, err := svc.Create("")
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
		if ve.Msg != "name is required" {
			t.Errorf("got msg %q, want %q", ve.Msg, "name is required")
		}
	})

	t.Run("valid name creates program", func(t *testing.T) {
		svc := newTestProgramService(t)

		p, err := svc.Create("Strength")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Name != "Strength" {
			t.Errorf("got name %q, want %q", p.Name, "Strength")
		}
	})
}

func TestProgramService_Update(t *testing.T) {
	t.Run("empty name returns ValidationError", func(t *testing.T) {
		svc := newTestProgramService(t)
		p, err := svc.Create("Strength")
		if err != nil {
			t.Fatal(err)
		}

		_, err = svc.Update(p.ID, "")
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})
}
