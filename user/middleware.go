package user

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/diegodario88/sesamo/config"
	"github.com/golang-jwt/jwt/v5"
)

type ContextKey string

type Checker interface {
	HasAccess(userID string, permission string) (bool, error)
}

const (
	UserIDKey    ContextKey = "userID"
	UserRolesKey ContextKey = "userRoles"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			http.Error(
				w,
				"Authorization header format must be Bearer {token}",
				http.StatusUnauthorized,
			)
			return
		}

		tokenStr := headerParts[1]
		claims := jwt.MapClaims{}

		token, err := jwt.ParseWithClaims(
			tokenStr,
			claims,
			func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}

				return []byte(config.Variables.JwtSecret), nil
			},
		)

		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		userID, ok := claims["userID"].(string)
		if !ok {
			http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)

		if roles, ok := claims["roles"].([]interface{}); ok {
			roleStrings := make([]string, len(roles))
			for i, role := range roles {
				roleStrings[i] = role.(string)
			}
			ctx = context.WithValue(ctx, UserRolesKey, roleStrings)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RBACMiddleware(svc Checker, permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(UserIDKey).(string)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			hasAccess, err := svc.HasAccess(userID, permission)
			if err != nil {
				http.Error(w, "Server error checking permissions", http.StatusInternalServerError)
				return
			}

			if !hasAccess {
				http.Error(w, "Permission denied", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
