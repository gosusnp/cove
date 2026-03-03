// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package httputil

import (
	"github.com/gosusnp/cove/backend/internal/domain"
	"net/http/httptest"
	"testing"
)

func TestMaskIP(t *testing.T) {
	tests := []struct {
		ip   string
		want string
	}{
		{"192.168.1.100", "192.168.1.0"},
		{"1.2.3.4", "1.2.3.0"},
		{"2001:db8:85a3:0:0:8a2e:370:7334", "2001:db8:85a3::"},
		{"invalid", ""},
		{"", ""},
	}

	for _, tt := range tests {
		got := maskIP(tt.ip)
		if got != domain.MaskedIP(tt.want) {
			t.Errorf("maskIP(%q) = %q, want %q", tt.ip, got, tt.want)
		}
	}
}

func TestRemoteIP(t *testing.T) {
	t.Run("X-Forwarded-For picks first entry", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("X-Forwarded-For", "203.0.113.1, 192.0.2.1")
		got := remoteIP(r)
		if got != "203.0.113.1" {
			t.Errorf("got %q, want %q", got, "203.0.113.1")
		}
	})

	t.Run("X-Forwarded-For trims leading space", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("X-Forwarded-For", " 203.0.113.1, 192.0.2.1")
		got := remoteIP(r)
		if got != "203.0.113.1" {
			t.Errorf("got %q, want %q", got, "203.0.113.1")
		}
	})

	t.Run("RemoteAddr fallback", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "192.0.2.1:1234"
		got := remoteIP(r)
		if got != "192.0.2.1" {
			t.Errorf("got %q, want %q", got, "192.0.2.1")
		}
	})
}
