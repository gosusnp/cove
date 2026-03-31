// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package middleware

import "net/http"

// RequireAdmin ensures the authenticated identity has is_admin = true.
// Returns 403 Forbidden otherwise. Must be applied after OAuth middleware.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := IdentityFromContext(r.Context())
		if !ok || !id.IsAdmin() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}
