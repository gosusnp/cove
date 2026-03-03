// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/httputil"
	"github.com/gosusnp/cove/backend/internal/service"
)

type identityCtxKey struct{}

// IdentityFromContext returns the identity stored in the request context.
func IdentityFromContext(ctx context.Context) (*domain.Identity, bool) {
	id, ok := ctx.Value(identityCtxKey{}).(*domain.Identity)
	return id, ok
}

// UserIDFromContext returns the authenticated user ID stored in the request context.
func UserIDFromContext(ctx context.Context) domain.UserID {
	if id, ok := IdentityFromContext(ctx); ok {
		return id.UserID
	}
	return domain.UserID{}
}

// OrgIDFromContext returns the org ID the auth token is scoped to.
func OrgIDFromContext(ctx context.Context) domain.OrgID {
	if id, ok := IdentityFromContext(ctx); ok {
		return id.OrgID
	}
	return domain.OrgID{}
}

// TokenIDFromContext returns the ID of the token used for authentication.
func TokenIDFromContext(ctx context.Context) uuid.UUID {
	if id, ok := IdentityFromContext(ctx); ok {
		return id.TokenID
	}
	return uuid.Nil
}

// OAuth guards routes by validating a session token via UserStore.
// On success, it stores the authenticated identity in the request context.
func OAuth(uSvc *service.UserService, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || token == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}
		ip, browser, os := httputil.FromRequest(r)
		user, org, tokenID, err := uSvc.GetUserByToken(r.Context(), token, ip, browser, os)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}

		id := &domain.Identity{
			UserID:  user.ID,
			OrgID:   org.ID,
			TokenID: tokenID,
		}

		ctx := context.WithValue(r.Context(), identityCtxKey{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
