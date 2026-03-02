package internal

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

const SHARED_SECRET = "a-string-secret-at-least-256-bits-long"

var SUBJECT = "subject"

func createToken(method jwt.SigningMethod, hmacSecret string, sub *string, roles []string) string {
	claims := authClaims{
		Roles: roles,
	}
	if sub != nil {
		claims.Subject = *sub
	}
	token := jwt.NewWithClaims(method, claims)

	var signed string
	var err error
	if method == jwt.SigningMethodNone {
		signed, err = token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	} else {
		signed, err = token.SignedString([]byte(hmacSecret))
	}
	if err != nil {
		panic(err)
	}
	return signed
}

func handlerReturningString(toReturn string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(toReturn))
	})
}

func (s *Server) handlerRequiringRole(role string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hasRole := s.hasRole(role, r)
		if !hasRole {
			http.Error(w, fmt.Sprintf("User does not have required role '%s'", role), http.StatusForbidden)
			return
		}
		w.Write([]byte("Authorized"))
	})
}

func createTestServer(requireAuth bool, hmacSecret string) *Server {
	router := http.NewServeMux()
	server := &Server{
		config: Config{
			RequireAuth: requireAuth,
			HmacSecret:  hmacSecret,
		},
		handler: router,
		logger:  slog.Default(),
	}
	router.Handle("/unauthenticated", handlerReturningString("No authentication required"))
	router.Handle("/authenticated", server.authMiddleware(handlerReturningString("Authenticated")))
	router.Handle("/authorized", server.authMiddleware(server.handlerRequiringRole("TestRole")))
	return server
}

func validateStatusCode(rr *httptest.ResponseRecorder, t *testing.T, expected int) {
	if status := rr.Code; status != expected {
		t.Errorf("handler returned wrong status code: got %v want %v", status, expected)
	}
}

func createGetRequest(path string, t *testing.T) *http.Request {
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		t.Fatal(err)
	}
	return req
}

func bearer(token string) string {
	return fmt.Sprintf("Bearer %s", token)
}

type authTestCase struct {
	Name          string
	Path          string
	Authorization string
	Expected      int
}

func testAuthedRequest(server *Server, tt authTestCase) func(*testing.T) {
	return func(t *testing.T) {
		req := createGetRequest(tt.Path, t)
		if tt.Authorization != "" {
			req.Header.Add("Authorization", tt.Authorization)
		}
		rr := httptest.NewRecorder()

		server.handler.ServeHTTP(rr, req)
		validateStatusCode(rr, t, tt.Expected)
	}
}

func TestNoAuthRequired(t *testing.T) {
	t.Parallel()

	testCases := []authTestCase{
		{
			Name:     "Authenticated",
			Path:     "/authenticated",
			Expected: http.StatusOK,
		},
		{
			Name:     "Authorized",
			Path:     "/authorized",
			Expected: http.StatusOK,
		},
	}

	server := createTestServer(false, "")
	for _, tt := range testCases {
		t.Run(tt.Name, testAuthedRequest(server, tt))
	}
}

func TestAuthentication(t *testing.T) {
	t.Parallel()

	testCases := []authTestCase{
		{
			Name:          "UnprotectedWithNoToken",
			Path:          "/unauthenticated",
			Authorization: "",
			Expected:      http.StatusOK,
		},
		{
			Name:          "AuthenticatesWithValidToken",
			Path:          "/authenticated",
			Authorization: bearer(createToken(jwt.SigningMethodHS256, SHARED_SECRET, &SUBJECT, []string{})),
			Expected:      http.StatusOK,
		},
		{
			Name:          "DoesNotAuthenticateWithNoHeader",
			Path:          "/authenticated",
			Authorization: "",
			Expected:      http.StatusUnauthorized,
		},
		{
			Name:          "DoesNotAuthenticateWithInvalidScheme",
			Path:          "/authenticated",
			Authorization: "InvalidScheme fakeToken",
			Expected:      http.StatusUnauthorized,
		},
		{
			Name:          "DoesNotAuthenticateWithoutSubject",
			Path:          "/authenticated",
			Authorization: bearer(createToken(jwt.SigningMethodHS256, SHARED_SECRET, nil, []string{})),
			Expected:      http.StatusUnauthorized,
		},
		{
			Name:          "DoesNotAuthenticateWithNoneJWTAlgo",
			Path:          "/authenticated",
			Authorization: bearer(createToken(jwt.SigningMethodNone, "", &SUBJECT, []string{})),
			Expected:      http.StatusUnauthorized,
		},
		{
			Name:          "DoesNotAuthenticateWithInvalidJWTAlgo",
			Path:          "/authenticated",
			Authorization: bearer(createToken(jwt.SigningMethodHS512, SHARED_SECRET, &SUBJECT, []string{})),
			Expected:      http.StatusUnauthorized,
		},
	}

	server := createTestServer(true, SHARED_SECRET)
	for _, tt := range testCases {
		t.Run(tt.Name, testAuthedRequest(server, tt))
	}
}

func TestAuthorization(t *testing.T) {
	t.Parallel()

	testCases := []authTestCase{
		{
			Name:          "AuthorizesWithValidRole",
			Path:          "/authorized",
			Authorization: bearer(createToken(jwt.SigningMethodHS256, SHARED_SECRET, &SUBJECT, []string{"TestRole"})),
			Expected:      http.StatusOK,
		},
		{
			Name:          "DoesNotAuthorizeWithInvalidRole",
			Path:          "/authorized",
			Authorization: bearer(createToken(jwt.SigningMethodHS256, SHARED_SECRET, &SUBJECT, []string{"InvalidRole"})),
			Expected:      http.StatusForbidden,
		},
		{
			Name:          "DoesNotAuthorizeWithNoRole",
			Path:          "/authorized",
			Authorization: bearer(createToken(jwt.SigningMethodHS256, SHARED_SECRET, &SUBJECT, []string{})),
			Expected:      http.StatusForbidden,
		},
	}

	server := createTestServer(true, SHARED_SECRET)
	for _, tt := range testCases {
		t.Run(tt.Name, testAuthedRequest(server, tt))
	}
}
