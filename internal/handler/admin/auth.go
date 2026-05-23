package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/texef-tech/winzzon-blog/internal/db/sqlc"
	"github.com/texef-tech/winzzon-blog/internal/handler"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	queries    *sqlc.Queries
	jwtSecret  string
	jwtExpiry  int
}

func NewAuthHandler(q *sqlc.Queries, jwtSecret string, jwtExpiry int) *AuthHandler {
	return &AuthHandler{queries: q, jwtSecret: jwtSecret, jwtExpiry: jwtExpiry}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.BadRequest(w, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		handler.BadRequest(w, "MISSING_FIELDS", "Username and password are required")
		return
	}

	admin, err := h.queries.GetAdminByUsername(r.Context(), req.Username)
	if err != nil {
		handler.Unauthorized(w, "INVALID_CREDENTIALS", "Invalid username or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		handler.Unauthorized(w, "INVALID_CREDENTIALS", "Invalid username or password")
		return
	}

	expiresAt := time.Now().Add(time.Duration(h.jwtExpiry) * time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": admin.ID.String(),
		"iat": time.Now().Unix(),
		"exp": expiresAt.Unix(),
	})

	tokenStr, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		handler.ServerError(w, "TOKEN_ERROR", "Failed to generate token")
		return
	}

	_ = h.queries.UpdateLastLogin(r.Context(), admin.ID)

	handler.JSON(w, http.StatusOK, loginResponse{
		Token:     tokenStr,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if len(authHeader) < 8 {
		handler.Unauthorized(w, "UNAUTHORIZED", "Missing token")
		return
	}

	tokenStr := authHeader[7:]
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(h.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		handler.Unauthorized(w, "UNAUTHORIZED", "Invalid token")
		return
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	sub, _ := claims.GetSubject()

	expiresAt := time.Now().Add(time.Duration(h.jwtExpiry) * time.Hour)
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": sub,
		"iat": time.Now().Unix(),
		"exp": expiresAt.Unix(),
	})

	newTokenStr, err := newToken.SignedString([]byte(h.jwtSecret))
	if err != nil {
		handler.ServerError(w, "TOKEN_ERROR", "Failed to refresh token")
		return
	}

	handler.JSON(w, http.StatusOK, loginResponse{
		Token:     newTokenStr,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}
