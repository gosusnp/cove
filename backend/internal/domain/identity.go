// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package domain

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Identity represents the authenticated caller (user, org, and specific token).
type Identity struct {
	UserID  UserID
	OrgID   OrgID
	TokenID uuid.UUID
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
	ID        UserID    `json:"id"`
	Email     Email     `json:"email"`
	GoogleSub GoogleSub `json:"-"`
	CreatedAt time.Time `json:"created_at"`
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
