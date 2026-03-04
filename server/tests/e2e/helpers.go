// Package e2e provides end-to-end testing infrastructure
package e2e

import (
	"net/url"
	"strconv"
	"strings"
)

// extractHost extracts host from database URL
func extractHost(dbURL string) string {
	u, err := url.Parse(dbURL)
	if err != nil {
		// Try parsing as DSN
		if strings.Contains(dbURL, "host=") {
			parts := strings.Split(dbURL, " ")
			for _, part := range parts {
				if strings.HasPrefix(part, "host=") {
					return strings.TrimPrefix(part, "host=")
				}
			}
		}
		return "localhost"
	}
	return u.Hostname()
}

// extractPort extracts port from database URL
func extractPort(dbURL string) int {
	u, err := url.Parse(dbURL)
	if err != nil {
		// Try parsing as DSN
		if strings.Contains(dbURL, "port=") {
			parts := strings.Split(dbURL, " ")
			for _, part := range parts {
				if strings.HasPrefix(part, "port=") {
					portStr := strings.TrimPrefix(part, "port=")
					if port, err := strconv.Atoi(portStr); err == nil {
						return port
					}
				}
			}
		}
		return 5432
	}
	port := u.Port()
	if port == "" {
		return 5432
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		return 5432
	}
	return p
}

// extractRedisHost extracts host from Redis address
func extractRedisHost(addr string) string {
	parts := strings.Split(addr, ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return "localhost"
}

// extractRedisPort extracts port from Redis address
func extractRedisPort(addr string) int {
	parts := strings.Split(addr, ":")
	if len(parts) > 1 {
		if port, err := strconv.Atoi(parts[1]); err == nil {
			return port
		}
	}
	return 6379
}
