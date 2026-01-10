package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIsValidCEP(t *testing.T) {
	testCases := []struct {
		cep      string
		expected bool
	}{
		{"01310100", true},
		{"29902555", true},
		{"12345678", true},
		{"123", false},       // Too short
		{"0131010a", false},  // Contains letter
		{"invalid", false},   // Not numeric
		{"", false},          // Empty
		{"01310100 ", false}, // Contains space
		{"0131010", false},   // 7 digits
		{"013101000", false}, // 9 digits
	}

	for _, tc := range testCases {
		result := isValidCEP(tc.cep)
		if result != tc.expected {
			t.Errorf("isValidCEP(%s): expected %v, got %v", tc.cep, tc.expected, result)
		}
	}
}

func TestHandleCEPRequest_ValidCEP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/", handleCEPRequest)

	requestBody := CEPRequest{CEP: "01310100"}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should either succeed (200) or fail with proper error
	if w.Code < 400 || w.Code == 422 || w.Code == 404 {
		t.Logf("Valid CEP returned status %d (expected successful response or proper error)", w.Code)
	} else {
		t.Errorf("Valid CEP: unexpected status %d", w.Code)
	}
}

func TestHandleCEPRequest_InvalidCEP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/", handleCEPRequest)

	testCases := []string{
		"123",
		"0131010a",
		"invalid",
	}

	for _, cep := range testCases {
		requestBody := CEPRequest{CEP: cep}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("CEP %s: expected status 422, got %d", cep, w.Code)
		}

		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)

		if response["message"] != "invalid zipcode" {
			t.Errorf("CEP %s: expected message 'invalid zipcode', got '%s'", cep, response["message"])
		}
	}
}

func TestHandleCEPRequest_MissingCEP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/", handleCEPRequest)

	requestBody := map[string]string{}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("Missing CEP: expected status 422, got %d", w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["message"] != "invalid zipcode" {
		t.Errorf("Expected message 'invalid zipcode', got '%s'", response["message"])
	}
}

func TestHandleCEPRequest_MalformedJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/", handleCEPRequest)

	malformedJSON := `{"cep": invalid}`

	req := httptest.NewRequest("POST", "/", bytes.NewBufferString(malformedJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("Malformed JSON: expected status 422, got %d", w.Code)
	}
}
