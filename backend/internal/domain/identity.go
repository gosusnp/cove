// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package domain

import (
	"github.com/google/uuid"
)

// Identity represents the authenticated caller (user, org, and specific token).
type Identity struct {
	UserID  UserID
	OrgID   OrgID
	TokenID uuid.UUID
}
