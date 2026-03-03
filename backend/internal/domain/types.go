// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package domain

import (
	"database/sql/driver"
	"time"

	"github.com/google/uuid"
)

type OrgID uuid.UUID

func NewOrgID() OrgID {
	return OrgID(uuid.Must(uuid.NewV7()))
}

func (id *OrgID) Scan(src any) error {
	return (*uuid.UUID)(id).Scan(src)
}

func (id OrgID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

type Org struct {
	ID        OrgID     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type UserID uuid.UUID

func (id *UserID) Scan(src any) error {
	return (*uuid.UUID)(id).Scan(src)
}

func (id UserID) Value() (driver.Value, error) {
	return uuid.UUID(id).Value()
}

type Email string

type GoogleSub string

func NewUserID() UserID {
	return UserID(uuid.Must(uuid.NewV7()))
}

type User struct {
	ID        UserID    `json:"id"`
	Email     Email     `json:"email"`
	GoogleSub GoogleSub `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}
