package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"mf935-telemetry/internal/auth"
	"mf935-telemetry/internal/poller"
)

func serveStatic() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
}

type loginRequest struct {
	Password string `json:"password"`
}

type loginResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
}

func loginHandler(session *auth.Session, p *poller.Poller, ctx context.Context) http.HandlerFunc {
	var started bool

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(loginResponse{OK: false, Message: "invalid request"})
			return
		}

		if err := session.Login(req.Password); err != nil {
			if isDeviceUnreachable(err.Error()) {
				// MiFi is down but the password may still be correct.
				// Return 503 so the client can skip the login modal.
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(loginResponse{OK: false, Message: "device unreachable"})
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(loginResponse{OK: false, Message: err.Error()})
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "mf_pw",
			Value:    req.Password,
			Path:     "/",
			Expires:  time.Now().Add(30 * 24 * time.Hour),
			HttpOnly: false,
			SameSite: http.SameSiteStrictMode,
		})

		if !started {
			started = true
			go p.Run(ctx)
		}

		json.NewEncoder(w).Encode(loginResponse{OK: true})
	}
}

func isDeviceUnreachable(msg string) bool {
	return strings.Contains(msg, "network is unreachable") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "no route to host") ||
		strings.Contains(msg, "i/o timeout") ||
		strings.Contains(msg, "dial tcp")
}
