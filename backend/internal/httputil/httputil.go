// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package httputil

import (
	"net"
	"net/http"
	"strings"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/mileusna/useragent"
)

// FromRequest extracts the masked client IP, browser name, and OS from an HTTP request.
func FromRequest(r *http.Request) (ipMasked domain.MaskedIP, browser string, os string) {
	ipMasked = maskIP(remoteIP(r))
	ua := useragent.Parse(r.UserAgent())
	return ipMasked, ua.Name, ua.OS
}

func remoteIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func maskIP(ipStr string) domain.MaskedIP {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return ""
	}
	if ip.To4() != nil {
		return domain.MaskedIP(ip.Mask(net.CIDRMask(24, 32)).String())
	}
	return domain.MaskedIP(ip.Mask(net.CIDRMask(64, 128)).String())
}
