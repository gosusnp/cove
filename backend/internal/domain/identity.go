// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package domain

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Identity represents the authenticated caller (user, org, and specific token).
type Identity struct {
	UserID         UserID
	OrgID          OrgID
	TokenID        uuid.UUID
	ServiceAccount bool
	Admin          bool
}

// IsServiceAccount reports whether this identity belongs to a service account
// rather than a regular user. Use this instead of checking ServiceAccount directly.
func (id *Identity) IsServiceAccount() bool {
	return id.ServiceAccount
}

// IsAdmin reports whether this identity has administrative privileges.
func (id *Identity) IsAdmin() bool {
	return id.Admin
}

type identityCtxKey struct{}

// NewContext returns a new context with the given identity.
func NewContext(ctx context.Context, id *Identity) context.Context {
	return context.WithValue(ctx, identityCtxKey{}, id)
}

// IdentityFromContext returns the identity stored in the request context.
func IdentityFromContext(ctx context.Context) (*Identity, bool) {
	id, ok := ctx.Value(identityCtxKey{}).(*Identity)
	return id, ok
}

// -----------------------------------------------------------------------------
// Organization
// -----------------------------------------------------------------------------

type OrgID ID[struct{ orgID struct{} }]

func NewOrgID() OrgID {
	return OrgID{UUID: uuid.Must(uuid.NewV7())}
}

type Org struct {
	ID        OrgID     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// -----------------------------------------------------------------------------
// User
// -----------------------------------------------------------------------------

type UserID ID[struct{ userID struct{} }]

func NewUserID() UserID {
	return UserID{UUID: uuid.Must(uuid.NewV7())}
}

type Email string

type GoogleSub string

type User struct {
	ID                UserID      `json:"id"`
	Email             Email       `json:"email"`
	GoogleSub         GoogleSub   `json:"-"`
	FitnessUnitSystem *UnitSystem `json:"fitness_unit_system,omitempty"`
	CookingUnitSystem *UnitSystem `json:"cooking_unit_system,omitempty"`
	CreatedAt         time.Time   `json:"created_at"`
	IsServiceAccount  bool        `json:"-"`
	IsAdmin           bool        `json:"-"`
}

// -----------------------------------------------------------------------------
// PAT
// -----------------------------------------------------------------------------

type TokenID ID[struct{ pat struct{} }]

func NewTokenID(u uuid.UUID) TokenID {
	return TokenID{UUID: u}
}

type PAT struct {
	ID         TokenID    `json:"id"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
}

// -----------------------------------------------------------------------------
// Session
// -----------------------------------------------------------------------------

type SessionID ID[struct{ session struct{} }]

func NewSessionID() SessionID {
	return SessionID{UUID: uuid.Must(uuid.NewV7())}
}

type MaskedIP string

type Session struct {
	ID              SessionID  `json:"id"`
	CreatedAt       time.Time  `json:"created_at"`
	LastUsedAt      *time.Time `json:"last_used_at"`
	InitialIPMasked *MaskedIP  `json:"initial_ip_masked,omitempty"`
	InitialBrowser  *string    `json:"initial_browser,omitempty"`
	InitialOS       *string    `json:"initial_os,omitempty"`
	LastIPMasked    *MaskedIP  `json:"last_ip_masked,omitempty"`
	LastBrowser     *string    `json:"last_browser,omitempty"`
	LastOS          *string    `json:"last_os,omitempty"`
}

// -----------------------------------------------------------------------------
// Hardened ID Helper
// -----------------------------------------------------------------------------

type ID[T any] struct {
	uuid.UUID
}

func (id *ID[T]) Scan(src any) error {
	return id.UUID.Scan(src)
}

func (id ID[T]) Value() (driver.Value, error) {
	return id.UUID.Value()
}

func (id ID[T]) String() string {
	return id.UUID.String()
}

func (id ID[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.UUID)
}

func (id *ID[T]) UnmarshalJSON(data []byte) error {
	var u uuid.UUID
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	id.UUID = u
	return nil
}

// -----------------------------------------------------------------------------
// Hardened Int64 ID Helper
// -----------------------------------------------------------------------------

type IntID[T any] int64

func (id IntID[T]) Int64() int64 {
	return int64(id)
}

func (id *IntID[T]) Scan(src any) error {
	if src == nil {
		return nil
	}
	switch v := src.(type) {
	case int64:
		*id = IntID[T](v)
		return nil
	default:
		return fmt.Errorf("cannot scan %T into %T", src, id)
	}
}

func (id IntID[T]) Value() (driver.Value, error) {
	return int64(id), nil
}

func (id IntID[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(id))
}

func (id *IntID[T]) UnmarshalJSON(data []byte) error {
	var i int64
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}
	*id = IntID[T](i)
	return nil
}
