// server_test.go
package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	cfg "github.com/gasparian/go-api-template/internal/config"
	strg "github.com/gasparian/go-api-template/pkg/storage"
)

// MockStorage is a mock implementation of the strg.Storage interface
type MockStorage struct {
	PingFunc func() error
	GetFunc  func(userID string) (strg.UserRecord, error)
	SetFunc  func(userID, itemID, visitorID string) error
}

func (m *MockStorage) Ping() error {
	return m.PingFunc()
}

func (m *MockStorage) Get(userID string) (strg.UserRecord, error) {
	return m.GetFunc(userID)
}

func (m *MockStorage) Set(userID, itemID, visitorID string) error {
	return m.SetFunc(userID, itemID, visitorID)
}

func setupApp(mockStorage strg.Storage) *App {
	config := cfg.ApplicationConfig{
		Version: "1.0.0",
		Name:    "ExampleApp",
	}
	allowCredentials := true
	corsConfig := cfg.CORSConfig{
		AllowedOrigins:   []string{"https://example.com"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowCredentials: &allowCredentials,
	}
	app := &App{}
	app.Initialize(config, corsConfig, mockStorage)

	// Replace logger with a no-op logger for testing
	logger, _ := zap.NewDevelopment()
	app.Logger = logger

	return app
}

func TestHandlePing_Success(t *testing.T) {
	mockStorage := &MockStorage{
		PingFunc: func() error {
			return nil
		},
	}
	app := setupApp(mockStorage)

	req, err := http.NewRequest("GET", "/internal/ping", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	expected := "pong"
	if rr.Body.String() != expected {
		t.Errorf("Expected body %q, got %q", expected, rr.Body.String())
	}
}

func TestHandlePing_Failure(t *testing.T) {
	mockStorage := &MockStorage{
		PingFunc: func() error {
			return errors.New("storage unreachable")
		},
	}
	app := setupApp(mockStorage)

	req, err := http.NewRequest("GET", "/internal/ping", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, status)
	}

	expected := "can't reach storage"
	if rr.Body.String() != expected {
		t.Errorf("Expected body %q, got %q", expected, rr.Body.String())
	}
}

func TestHandleVersion(t *testing.T) {
	mockStorage := &MockStorage{}
	app := setupApp(mockStorage)

	req, err := http.NewRequest("GET", "/internal/version", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	expectedVersion := "1.0.0"
	if response["version"] != expectedVersion {
		t.Errorf("Expected version %q, got %q", expectedVersion, response["version"])
	}
}

func TestHandleUsersGet_Success(t *testing.T) {
	mockStorage := &MockStorage{
		GetFunc: func(userID string) (strg.UserRecord, error) {
			return strg.UserRecord{
				UserID:         userID,
				TotalItemsSeen: 5,
				LastItemID:     "item-123",
			}, nil
		},
	}
	app := setupApp(mockStorage)

	req, err := http.NewRequest("GET", "/api/v1/users?user_id=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	var response GetResponseModel
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.UserID != "123" {
		t.Errorf("Expected UserID %q, got %q", "123", response.UserID)
	}
	if response.TotalItemsSeen != 5 {
		t.Errorf("Expected TotalItemsSeen %d, got %d", 5, response.TotalItemsSeen)
	}
	if response.LastItemID != "item-123" {
		t.Errorf("Expected LastItemID %q, got %q", "item-123", response.LastItemID)
	}
}

func TestHandleUsersGet_MissingUserID(t *testing.T) {
	mockStorage := &MockStorage{}
	app := setupApp(mockStorage)

	req, err := http.NewRequest("GET", "/api/v1/users", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, status)
	}

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	expectedError := "Missing user_id parameter"
	if response["error"] != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, response["error"])
	}
}

func TestHandleUsersGet_StorageError(t *testing.T) {
	mockStorage := &MockStorage{
		GetFunc: func(userID string) (strg.UserRecord, error) {
			return strg.UserRecord{}, errors.New("storage error")
		},
	}
	app := setupApp(mockStorage)

	req, err := http.NewRequest("GET", "/api/v1/users?user_id=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, status)
	}

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	expectedError := "Failed to retrieve user data"
	if response["error"] != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, response["error"])
	}
}

func TestHandleUsersPost_Success(t *testing.T) {
	mockStorage := &MockStorage{
		SetFunc: func(userID, itemID, visitorID string) error {
			return nil
		},
	}
	app := setupApp(mockStorage)

	payload := PostRequestModel{
		UserID: "123",
		ItemID: "item-456",
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(&http.Cookie{
		Name:  visitorCookieName,
		Value: "visitor-123",
	})
	rr := httptest.NewRecorder()

	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, status)
	}
}

func TestHandleUsersPost_InvalidPayload(t *testing.T) {
	mockStorage := &MockStorage{}
	app := setupApp(mockStorage)

	invalidJSON := `{"userId": "123", "itemId": "item-456"` // Missing closing brace

	req, err := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer([]byte(invalidJSON)))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, status)
	}

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	expectedError := "Invalid request payload"
	if response["error"] != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, response["error"])
	}
}

func TestHandleUsersPost_StorageError(t *testing.T) {
	mockStorage := &MockStorage{
		SetFunc: func(userID, itemID, visitorID string) error {
			return errors.New("storage set error")
		},
	}
	app := setupApp(mockStorage)

	payload := PostRequestModel{
		UserID: "123",
		ItemID: "item-456",
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, status)
	}
}

func TestLoggingMiddleware(t *testing.T) {
	// This test ensures that the logging middleware does not interfere with request handling.
	// Since logging is internal, we'll just ensure that the request is handled correctly.

	mockStorage := &MockStorage{}
	app := setupApp(mockStorage)

	// Create a simple handler to test middleware
	app.Router.HandleFunc("/test/middleware", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("middleware test"))
	}).Methods("GET")

	req, err := http.NewRequest("GET", "/test/middleware", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Referer", "http://example.com")
	rr := httptest.NewRecorder()

	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	expected := "middleware test"
	if rr.Body.String() != expected {
		t.Errorf("Expected body %q, got %q", expected, rr.Body.String())
	}
}

func TestHandleUsersPost_NoAnonymousCookie(t *testing.T) {
	mockStorage := &MockStorage{
		SetFunc: func(userID, itemID, visitorID string) error {
			if visitorID != "" {
				return errors.New("expected empty visitorID")
			}
			return nil
		},
	}
	app := setupApp(mockStorage)

	payload := PostRequestModel{
		UserID: "123",
		ItemID: "item-456",
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	// Do not add the visitor cookie
	rr := httptest.NewRecorder()

	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, status)
	}
}

func TestHandleUsersPost_WithAnonymousCookie(t *testing.T) {
	mockStorage := &MockStorage{
		SetFunc: func(userID, itemID, visitorID string) error {
			if visitorID != "visitor-123" {
				return errors.New("expected visitorID 'visitor-123'")
			}
			return nil
		},
	}
	app := setupApp(mockStorage)

	payload := PostRequestModel{
		UserID: "123",
		ItemID: "item-456",
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(&http.Cookie{
		Name:  visitorCookieName,
		Value: "visitor-123",
	})
	rr := httptest.NewRecorder()

	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, status)
	}
}

func TestHandleUsersPost_InvalidMethod(t *testing.T) {
	mockStorage := &MockStorage{}
	app := setupApp(mockStorage)

	req, err := http.NewRequest("PUT", "/api/v1/users", nil) // PUT is not allowed
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, status)
	}
}

func TestCORSHeaders(t *testing.T) {
	mockStorage := &MockStorage{}
	app := setupApp(mockStorage)

	req, err := http.NewRequest("OPTIONS", "/api/v1/users", nil)
	if err != nil {
		t.Fatal(err)
	}
	origin := "https://example.com"
	expectedMethods := "POST"
	req.Header.Add("Origin", origin)
	req.Header.Add("Access-Control-Request-Method", expectedMethods)

	rr := httptest.NewRecorder()

	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	// Check CORS headers
	allowedMethods := rr.Header().Get("Access-Control-Allow-Methods")
	if allowedMethods != expectedMethods {
		t.Errorf("Expected Access-Control-Allow-Methods %q, got %q", expectedMethods, allowedMethods)
	}

	allowedOrigins := rr.Header().Get("Access-Control-Allow-Origin")
	if allowedOrigins != origin {
		t.Errorf("Expected Access-Control-Allow-Origin %q, got %q", allowedOrigins, origin)
	}
}

func TestHandleUsersPost_EmptyBody(t *testing.T) {
	mockStorage := &MockStorage{}
	app := setupApp(mockStorage)

	req, err := http.NewRequest("POST", "/api/v1/users", nil) // Empty body
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, status)
	}

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	expectedError := "Invalid request payload"
	if response["error"] != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, response["error"])
	}
}

func TestHandleUsersGet_InvalidQueryParam(t *testing.T) {
	mockStorage := &MockStorage{}
	app := setupApp(mockStorage)

	// user_id is present but empty
	req, err := http.NewRequest("GET", "/api/v1/users?user_id=", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, status)
	}

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	expectedError := "Missing user_id parameter"
	if response["error"] != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, response["error"])
	}
}
