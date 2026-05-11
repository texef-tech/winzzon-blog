package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const AdminIDKey contextKey = "admin_id"

func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"Missing authorization header"}}`, http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenStr == authHeader {
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"Invalid authorization format"}}`, http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"Invalid or expired token"}}`, http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"Invalid token claims"}}`, http.StatusUnauthorized)
				return
			}

			sub, _ := claims.GetSubject()
			adminID, err := uuid.Parse(sub)
			if err != nil {
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"Invalid token subject"}}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), AdminIDKey, adminID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetAdminID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(AdminIDKey).(uuid.UUID)
	return id, ok
}
