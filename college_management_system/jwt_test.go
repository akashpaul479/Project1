package collegemanagementsystem_test

import (
	"bytes"
	collegemanagementsystem "college_management_system/college_management_system"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAccessToken(t *testing.T) {
	collegemanagementsystem.SecretKey = []byte("test_secret")

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		email    string
		want     string
		willpass bool
	}{
		{
			name:     "valid email",
			email:    "akash@gmail.com",
			willpass: false,
		},
		{
			name:     "empty email",
			email:    "",
			willpass: false, // your function does NOT validate email
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := collegemanagementsystem.GenerateAccessToken(tt.email)
			if err != nil {
				if !tt.willpass {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if tt.willpass {
				t.Fatal("Expected error got nil")
			}
			// Token should not be empty
			if got == "" {
				t.Fatal("Empty token generated")
			}
			// Validate token
			claims := &collegemanagementsystem.Claims{}

			token, err := jwt.ParseWithClaims(got, claims, func(t *jwt.Token) (interface{}, error) {
				return collegemanagementsystem.SecretKey, nil
			})

			if err != nil || !token.Valid {
				t.Fatalf("Generated token is invalid!")
			}
			if claims.Email != tt.email {
				t.Fatalf("Expected email %s , got %s", tt.email, claims.Email)
			}
			if claims.TokenType != "access" {
				t.Fatalf("expected access got %s ", claims.TokenType)
			}
		})
	}
}

func TestGenerateRefreshToken(t *testing.T) {

	collegemanagementsystem.SecretKey = []byte("test_secret")

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		email    string
		want     string
		willpass bool
	}{
		{
			name:     "valid email",
			email:    "akash@gmail.com",
			willpass: false,
		},
		{
			name:     "empty email",
			email:    "",
			willpass: false, // your function does NOT validate email
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := collegemanagementsystem.GenerateRefreshToken(tt.email)
			if err != nil {
				if !tt.willpass {
					t.Fatal(err)
				}
				return
			}
			if got == "" {
				t.Fatal("empty token")
			}
			claims := &collegemanagementsystem.Claims{}

			token, err := jwt.ParseWithClaims(got, claims, func(t *jwt.Token) (interface{}, error) {
				return collegemanagementsystem.SecretKey, nil
			})

			if err != nil || !token.Valid {
				t.Fatal("invalid token")
			}

			if claims.TokenType != "refresh" {
				t.Fatalf("not refresh token")
			}
		})
	}
}

func TestSetAccessCookies(t *testing.T) {
	tests := []struct {
		name  string // description of this test case
		token string
	}{
		{
			name:  "valid token",
			token: "access-token-123",
		},
		{
			name:  "empty token",
			token: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			w := httptest.NewRecorder()
			collegemanagementsystem.SetAccessCookies(w, tt.token)

			r := w.Result()
			cookies := r.Cookies()

			if len(cookies) == 0 {
				t.Fatal("no cookies set")
			}
			var access *http.Cookie

			for _, c := range cookies {
				if c.Name == "access_token" {
					access = c
				}

			}
			if access == nil {
				t.Fatalf("access_token cookie not found")
			}
			if access.Value != tt.token {
				t.Fatalf("expected %s , got %s", tt.token, access.Value)
			}
			if !access.HttpOnly {
				t.Fatal("cookie must be HTTP only")
			}
			if access.Path != "/" {
				t.Fatal("cookie path must be /")
			}
			if access.Expires.Before(time.Now()) {
				t.Fatal("cookie already expired")
			}
		})
	}
}

func TestSetRefreshCookies(t *testing.T) {
	tests := []struct {
		name  string // description of this test case
		token string
	}{
		{
			name:  "valid token",
			token: "refresh-token-123",
		},
		{
			name:  "invalid token",
			token: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			w := httptest.NewRecorder()
			collegemanagementsystem.SetRefreshCookies(w, tt.token)

			r := w.Result()
			cookies := r.Cookies()

			var refresh *http.Cookie

			for _, c := range cookies {
				refresh = c
			}

			if refresh == nil {
				t.Fatal("refresh_token cookie not found")
			}

			if refresh.Value != tt.token {
				t.Fatalf("Expected %s , got %s", tt.token, refresh.Value)
			}

			if !refresh.HttpOnly {
				t.Fatal("Cookie must be http only")
			}
			if refresh.Path != "/" {
				t.Fatal("cookie path must be /")
			}
			if refresh.Expires.Before(time.Now()) {
				t.Fatal("cookie already expired")
			}
		})
	}
}

func TestClearAccessCookies(t *testing.T) {
	tests := []struct {
		name  string // description of this test case
		token string
	}{
		{
			name:  "valid token",
			token: "refresh-token-123",
		},
		{
			name:  "invalid token",
			token: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			w := httptest.NewRecorder()
			collegemanagementsystem.ClearAccessCookies(w)

			r := w.Result()
			cookies := r.Cookies()

			var access *http.Cookie

			for _, c := range cookies {
				if c.Name == "access_token" {
					access = c
				}
			}
			if access == nil {
				t.Fatal("access_token cookie not found")
			}
			if access.Value != "" {
				t.Fatal("cookie value not cleared")
			}

			if access.Expires.After(time.Now()) {
				t.Fatal("cookie not expired")
			}
		})
	}
}

func TestLoginHandler(t *testing.T) {

	// Setup env
	os.Setenv("EMAIL", "akashpaul@gmail.com")
	os.Setenv("PASSWORD", "Akash@479")

	// Setup JWT secret
	collegemanagementsystem.SecretKey = []byte("test-secret")

	tests := []struct {
		name     string // description of this test case
		email    string
		password string
		willpass bool
	}{
		{
			name:     "valid login",
			email:    "akashpaul@gmail.com",
			password: "Akash@479",
			willpass: true,
		},
		{
			name:     "invalid email",
			email:    "",
			password: "Akash@479",
			willpass: false,
		},
		{
			name:     "invalid password",
			email:    "akashpaul@gmail.com",
			password: "",
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			userBody, err := json.Marshal(map[string]string{"email": tt.email, "password": tt.password})
			if err != nil {
				panic(err)
			}
			buffer := bytes.NewBuffer(userBody)
			r := httptest.NewRequest(http.MethodPost, "/login", buffer)
			w := httptest.NewRecorder()

			collegemanagementsystem.LoginHandler(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected OK status , got %d", w.Code)
				}

				cookies := w.Result().Cookies()

				if len(cookies) < 2 {
					t.Fatal("expected access and refresh cookies")
				}
				foundaccess := false
				foundrefresh := false

				for _, c := range cookies {
					if c.Name == "access_token" {
						foundaccess = true
					}
					if c.Name == "refresh_token" {
						foundrefresh = true
					}
				}
				if !foundaccess || !foundrefresh {
					t.Fatal("missing jwt cookies")
				}
				if w.Code == http.StatusUnauthorized {
					t.Fatalf("Expected not ok status , got %d", w.Code)
				}
			}
		})
	}
}

func TestRefreshHandler(t *testing.T) {

	// Setup env
	os.Setenv("EMAIL", "akashpaul@gmail.com")
	os.Setenv("PASSWORD", "Akash@479")

	// Setup JWT secret
	collegemanagementsystem.SecretKey = []byte("test-secret")

	tests := []struct {
		name      string // description of this test case
		tokenType string
		setCookie bool
		willpass  bool
	}{
		{
			name:      "valid refresh token",
			tokenType: "refresh",
			setCookie: true,
			willpass:  true,
		},
		{
			name:      "missing refresh token",
			tokenType: "",
			setCookie: false,
			willpass:  false,
		},
		{
			name:      "invalid token type",
			tokenType: "access",
			setCookie: true,
			willpass:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := httptest.NewRequest(http.MethodPost, "/refresh", nil)

			if tt.setCookie {
				claims := collegemanagementsystem.Claims{
					Email:     "akashpaul@gmail.com",
					TokenType: tt.tokenType,
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

				tokenstr, _ := token.SignedString(collegemanagementsystem.SecretKey)

				r.AddCookie(&http.Cookie{Name: "refresh_token", Value: tokenstr})
			}
			w := httptest.NewRecorder()
			collegemanagementsystem.RefreshHandler(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected Ok status , got %d", w.Code)
				}
				cookies := w.Result().Cookies()

				found := false

				for _, c := range cookies {
					if c.Name == "access_token" {
						found = true
					}
					if !found {
						t.Fatal("access token cookie not set")
					}
				}
			}
		})
	}
}

func TestJwtMiddleware(t *testing.T) {

	// Setup env
	os.Setenv("EMAIL", "akashpaul@gmail.com")
	os.Setenv("PASSWORD", "Akash@479")

	// Setup JWT secret
	collegemanagementsystem.SecretKey = []byte("test-secret")

	tests := []struct {
		name      string // description of this test case
		tokenType string
		setCookie bool
		willpass  bool
	}{
		{
			name:      "valid access token",
			tokenType: "access",
			setCookie: true,
			willpass:  true,
		},
		{
			name:      "missing token",
			tokenType: "",
			setCookie: false,
			willpass:  false,
		},
		{
			name:      "invalid token type",
			tokenType: "refresh",
			setCookie: true,
			willpass:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			Protected := http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					email := r.Header.Get("X-User-Email")

					if email != "akashpaul@gmail.com" {
						t.Fatalf("Wrong user email: %s", email)
					}
				},
			)

			collegemanagementsystem.JwtMiddleware(Protected)

			r := httptest.NewRequest(http.MethodGet, "/test", nil)

			if tt.willpass {
				claims := &collegemanagementsystem.Claims{
					Email:     "akashpaul@gmail.com",
					TokenType: tt.tokenType,
				}

				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

				tokenstr, _ := token.SignedString(collegemanagementsystem.SecretKey)

				r.AddCookie(&http.Cookie{
					Name:  "access_token",
					Value: tokenstr,
				})
			}
			w := httptest.NewRecorder()

			if w.Code != http.StatusOK {
				t.Fatalf("Expected not OK , got %d", w.Code)
			}

		})
	}
}

func TestLogoutHandler(t *testing.T) {

	// Setup env
	os.Setenv("EMAIL", "akashpaul@gmail.com")
	os.Setenv("PASSWORD", "Akash@479")

	// Setup JWT secret
	collegemanagementsystem.SecretKey = []byte("test-secret")

	tests := []struct {
		name      string // description of this test case
		tokenType string
		setCookie bool
		willpass  bool
	}{
		{
			name:      "valid access token",
			tokenType: "access",
			setCookie: true,
			willpass:  true,
		},
		{
			name:      "missing token",
			tokenType: "",
			setCookie: false,
			willpass:  false,
		},
		{
			name:      "invalid token type",
			tokenType: "refresh",
			setCookie: true,
			willpass:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/logout", nil)
			w := httptest.NewRecorder()

			collegemanagementsystem.LogoutHandler(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected OK status , got %d", w.Code)
				}
				cookies := w.Result().Cookies()

				var access *http.Cookie

				for _, c := range cookies {
					if c.Name == "access_token" {
						access = c
					}
					if access == nil {
						t.Fatal("access_token choice not found")
					}
					if access.Value != "" {
						t.Fatal("access token not cleared")
					}
					if access.Expires.After(time.Now()) {
						t.Fatal("cookie not expired")
					}
				}
			}
		})
	}
}
