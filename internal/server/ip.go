package server

import (
	"net"
	"net/http"
	"strings"
)

// GetClientIP attempts to determine the client's real IP address.
// It checks X-Forwarded-For and X-Real-IP headers, falling back to RemoteAddr.
func getClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		// The first IP in the list is the client's IP
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		ip := strings.TrimSpace(xRealIP)
		if net.ParseIP(ip) != nil {
			return ip
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	if net.ParseIP(ip) != nil {
		return ip
	}

	return ""
}
