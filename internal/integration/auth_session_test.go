package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"drone-operations-control/internal/console"
	"drone-operations-control/internal/httpapi"
	"drone-operations-control/internal/repository"
	"drone-operations-control/internal/service"
)

func TestConsoleSessionLoginInfoLogoutAndRevocation(t *testing.T) {
	pool, _ := openDatabase(t)
	defer pool.Close()
	handler := httpapi.New(service.New(repository.NewPostgres(pool)), pool.Ping).WithConsole(console.NewStore(pool)).Handler()

	loginBody := bytes.NewBufferString(`{"Username":"operator","Password":"operator123"}`)
	loginRequest := httptest.NewRequest(http.MethodPost, "/api/auth/login", loginBody)
	loginResponse := httptest.NewRecorder()
	handler.ServeHTTP(loginResponse, loginRequest)
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", loginResponse.Code, loginResponse.Body.String())
	}
	var envelope struct {
		Data struct {
			Token string       `json:"token"`
			User  console.User `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(loginResponse.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data.Token == "" || envelope.Data.User.Username != "operator" || envelope.Data.User.Role != 0 {
		t.Fatalf("login data=%+v", envelope.Data)
	}

	infoRequest := httptest.NewRequest(http.MethodGet, "/api/auth/info", nil)
	infoRequest.Header.Set("Authorization", "Bearer "+envelope.Data.Token)
	infoResponse := httptest.NewRecorder()
	handler.ServeHTTP(infoResponse, infoRequest)
	if infoResponse.Code != http.StatusOK {
		t.Fatalf("info status=%d body=%s", infoResponse.Code, infoResponse.Body.String())
	}

	logoutRequest := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	logoutRequest.Header.Set("Authorization", "Bearer "+envelope.Data.Token)
	logoutResponse := httptest.NewRecorder()
	handler.ServeHTTP(logoutResponse, logoutRequest)
	if logoutResponse.Code != http.StatusOK {
		t.Fatalf("logout status=%d body=%s", logoutResponse.Code, logoutResponse.Body.String())
	}

	revokedRequest := httptest.NewRequest(http.MethodGet, "/api/auth/info", nil)
	revokedRequest.Header.Set("Authorization", "Bearer "+envelope.Data.Token)
	revokedResponse := httptest.NewRecorder()
	handler.ServeHTTP(revokedResponse, revokedRequest)
	if revokedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("revoked status=%d body=%s", revokedResponse.Code, revokedResponse.Body.String())
	}
}

func TestConsoleExpiredSessionIsRejected(t *testing.T) {
	pool, _ := openDatabase(t)
	defer pool.Close()
	store := console.NewStore(pool)
	var userID string
	if err := pool.QueryRow(t.Context(), `SELECT id FROM console_users WHERE username='admin'`).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	token := "90000000-0000-0000-0000-000000000001"
	if _, err := pool.Exec(t.Context(), `INSERT INTO console_sessions(token,user_id,expires_at) VALUES($1,$2,now()-interval '1 minute') ON CONFLICT(token) DO UPDATE SET expires_at=excluded.expires_at,revoked_at=NULL`, token, userID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SessionUser(t.Context(), token); err == nil {
		t.Fatal("expired session was accepted")
	}
}
