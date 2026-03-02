// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/gosusnp/cove/backend/internal/httputil"
	"github.com/gosusnp/cove/backend/internal/store"
)

type userCtxKey struct{}
type orgCtxKey struct{}
type tokenIDCtxKey struct{}

// UserFromContext returns the authenticated user stored in the request context by OAuth middleware.
func UserFromContext(ctx context.Context) *store.User {
	u, _ := ctx.Value(userCtxKey{}).(*store.User)
	return u
}

// OrgFromContext returns the org the auth token is scoped to.
func OrgFromContext(ctx context.Context) *store.Org {
	o, _ := ctx.Value(orgCtxKey{}).(*store.Org)
	return o
}

// TokenIDFromContext returns the ID of the token used for authentication.
func TokenIDFromContext(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(tokenIDCtxKey{}).(uuid.UUID)
	return id
}

// OAuth guards routes by validating a session token via UserStore.
// On success, it stores the authenticated user and org in the request context.
func OAuth(us *store.UserStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || token == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}
		ip, browser, os := httputil.FromRequest(r)
		user, org, tokenID, err := us.GetUserByToken(token, ip, browser, os)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}
		ctx := context.WithValue(r.Context(), userCtxKey{}, user)
		ctx = context.WithValue(ctx, orgCtxKey{}, org)
		ctx = context.WithValue(ctx, tokenIDCtxKey{}, tokenID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
