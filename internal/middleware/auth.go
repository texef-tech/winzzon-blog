package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/texef-tech/winzzon-blog/internal/handler"
)

type contextKey string

const AdminIDKey contextKey = "admin_id"

func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				handler.Unauthorized(w, "UNAUTHORIZED", "Missing authorization header")
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenStr == authHeader {
				handler.Unauthorized(w, "UNAUTHORIZED", "Invalid authorization format")
				return
			}

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				handler.Unauthorized(w, "UNAUTHORIZED", "Invalid or expired token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				handler.Unauthorized(w, "UNAUTHORIZED", "Invalid token claims")
				return
			}

			sub, _ := claims.GetSubject()
			adminID, err := uuid.Parse(sub)
			if err != nil {
				handler.Unauthorized(w, "UNAUTHORIZED", "Invalid token subject")
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
