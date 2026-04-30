package auth

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/qaynaq/qaynaq/internal/config"
)

type BasicAuthHandler struct {
	username   string
	password   string
	jwtManager *JWTManager
	cookieName string
}

func NewBasicAuthHandler(cfg *config.AuthConfig, secretKey string) *BasicAuthHandler {
	return &BasicAuthHandler{
		username:   cfg.BasicUsername,
		password:   cfg.BasicPassword,
		jwtManager: NewJWTManager(secretKey, 24*time.Hour),
		cookieName: "qaynaq_session",
	}
}

func (h *BasicAuthHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := authHeader[7:]
			claims, err := h.jwtManager.ValidateToken(token)
			if err == nil && claims.AuthType == "basic" {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error":"Unauthorized"}`))
	})
}

func (h *BasicAuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var loginReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	usernameMatch := subtle.ConstantTimeCompare([]byte(loginReq.Username), []byte(h.username)) == 1
	passwordMatch := subtle.ConstantTimeCompare([]byte(loginReq.Password), []byte(h.password)) == 1

	if !usernameMatch || !passwordMatch {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := h.jwtManager.GenerateToken(loginReq.Username, "", "basic")
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate JWT token")
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	log.Info().Str("username", loginReq.Username).Msg("User logged in with basic auth")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token":      token,
		"token_type": "Bearer",
	})
}

func (h *BasicAuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Clear the cookie set by /auth/exchange.
	http.SetCookie(w, &http.Cookie{
		Name:     h.cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *BasicAuthHandler) isValidSession(token string) bool {
	claims, err := h.jwtManager.ValidateToken(token)
	if err != nil {
		return false
	}
	return claims.AuthType == "basic"
}
