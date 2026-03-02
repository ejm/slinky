package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
	sloghttp "github.com/samber/slog-http"
)

type contextKey string

const tokenKey contextKey = "token"

type authClaims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

func (a authClaims) Validate() error {
	subject, err := a.GetSubject()
	if err != nil {
		return err
	}
	if subject == "" {
		return errors.New("'sub' is a required claim")
	}
	return nil
}

// Parse a JWT string with the custom claims and Keyfunc
func (s *Server) parseJwtToken(tokenString string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(
		tokenString,
		&authClaims{},
		func(t *jwt.Token) (any, error) {
			return []byte(s.config.HmacSecret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
}

// Check if a request's token has the given role
func (s *Server) hasRole(role string, r *http.Request) bool {
	if s.config.RequireAuth == false {
		return true
	}
	token, ok := r.Context().Value(tokenKey).(*jwt.Token)
	// This function should only ever be run from a Handler that has
	// gone through `authMiddleware`, so these two checks *should* always be true
	if !ok {
		s.logger.ErrorContext(r.Context(), "Request doesn't have an attached token")
		return false
	}
	claims, ok := token.Claims.(*authClaims)
	if !ok {
		s.logger.ErrorContext(r.Context(), "Request token has invalid claims")
		return false
	}
	return slices.Contains(claims.Roles, role)
}

// Middleware to accept a JWT token and include it in the request context
// If the JWT token isn't valid, the `next` Handler is never called
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.config.RequireAuth == false {
			next.ServeHTTP(w, r)
			return
		}
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}
		tokenString, isBearerToken := strings.CutPrefix(authHeader, "Bearer ")
		if !isBearerToken {
			http.Error(w, "Authorization header scheme unsupported", http.StatusUnauthorized)
			return
		}
		token, err := s.parseJwtToken(tokenString)
		if err != nil {
			s.logger.ErrorContext(r.Context(), err.Error())
			http.Error(w, fmt.Sprintf("Error parsing JWT token: %s", err.Error()), http.StatusInternalServerError)
		} else if claims, ok := token.Claims.(*authClaims); ok {
			sloghttp.AddCustomAttributes(
				r,
				slog.String("token.sub", claims.Subject),
				slog.Any("token.roles", claims.Roles),
			)
			ctx := context.WithValue(r.Context(), tokenKey, token)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			s.logger.ErrorContext(r.Context(), "Unknown claims type, cannot proceed")
			http.Error(w, "Invalid claims on supplied JWT", http.StatusUnauthorized)
		}
	})
}
