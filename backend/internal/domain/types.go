// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrgID uuid.UUID

func NewOrgID() OrgID {
	return OrgID(uuid.Must(uuid.NewV7()))
}

type Org struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type UserID uuid.UUID
