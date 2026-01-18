/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

package middleware

import (
	"net/http"
	"time"

	"github.com/steveredden/KindredCard/internal/logger"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// LoggingMiddleware logs all HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Process request
		next.ServeHTTP(wrapped, r)

		// Log after request completes
		duration := time.Since(start)

		logger.Info("[WEB] [%s] %s %s - Status: %d - Duration: %v - Size: %d bytes - IP: %s",
			r.Method,
			r.URL.Path,
			r.Proto,
			wrapped.statusCode,
			duration,
			wrapped.written,
			r.RemoteAddr,
		)

		// Log user if authenticated (from context)
		if user, ok := GetUserFromContext(r); ok {
			logger.Debug("[WEB] ↳ User: %s (ID: %d)", user.Email, user.ID)
		}

		// Log any errors (4xx, 5xx)
		if wrapped.statusCode >= 400 {
			logger.Error("[WEB] ↳ Response: %d %s for %s %s",
				wrapped.statusCode,
				http.StatusText(wrapped.statusCode),
				r.Method,
				r.URL.Path,
			)
		}
	})
}
