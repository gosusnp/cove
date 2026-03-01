// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"errors"
	"testing"

	"github.com/gosusnp/cove/backend/internal/store"
)

func newTestUserService(t *testing.T) (*UserService, *store.UserStore) {
	t.Helper()
	us := store.NewUserStore(newTestDB(t))
	return NewUserService(us), us
}

func TestUserService_Get(t *testing.T) {
	t.Run("returns existing user", func(t *testing.T) {
		svc, us := newTestUserService(t)

		created, _, err := us.GetOrCreate("get@example.com", "sub-get")
		if err != nil {
			t.Fatalf("GetOrCreate: %v", err)
		}

		got, err := svc.Get(created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Email != "get@example.com" {
			t.Errorf("got email %q, want %q", got.Email, "get@example.com")
		}
		if got.ID != created.ID {
			t.Errorf("got id %v, want %v", got.ID, created.ID)
		}
	})

	t.Run("unknown id returns ErrNotFound", func(t *testing.T) {
		svc, us := newTestUserService(t)

		// Create a user to generate a valid UUID format, then discard it.
		user, _, err := us.GetOrCreate("other@example.com", "sub-other")
		if err != nil {
			t.Fatalf("GetOrCreate: %v", err)
		}

		// Flip one byte so the UUID is valid but unknown.
		id := user.ID
		id[0] ^= 0xff

		_, err = svc.Get(id)
		if !errors.Is(err, store.ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}
