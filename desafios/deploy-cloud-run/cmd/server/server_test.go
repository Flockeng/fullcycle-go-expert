package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// Helper function to set environment variables in tests
func setEnv(key, value string) func() {
	old := os.Getenv(key)
	os.Setenv(key, value)
	return func() { os.Setenv(key, old) }
}

// Helper function to mock http.Get
func patchHTTPGet(fn func(url string) (*http.Response, error)) func() {
	orig := httpGet
	httpGet = fn
	return func() { httpGet = orig }
}

// TestIsValidString tests the isValidString helper function
func TestIsValidString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Non-empty string", "São Paulo", true},
		{"String with spaces", "   São Paulo   ", true},
		{"Empty string", "", false},
		{"Only spaces", "   ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidString(tt.input)
			if result != tt.expected {
				t.Errorf("isValidString(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestRound tests the round helper function
func TestRound(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		precision int
		expected  float64
	}{
		{"Round to 1 decimal", 25.567, 1, 25.6},
		{"Round to 2 decimals", 25.567, 2, 25.57},
		{"Round to 0 decimals", 25.567, 0, 26.0},
		{"Negative number", -15.456, 1, -15.5},
		{"Already rounded", 25.5, 1, 25.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := round(tt.value, tt.precision)
			if result != tt.expected {
				t.Errorf("round(%f, %d) = %f, want %f", tt.value, tt.precision, result, tt.expected)
			}
		})
	}
}

// TestGetWeatherByZipCode_InvalidZipcode tests invalid zipcode format
func TestGetWeatherByZipCode_InvalidZipcode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/zipweather/:zipcode", getWeatherByZipCode)

	tests := []struct {
		name    string
		zipcode string
	}{
		{"Short zipcode", "123"},
		{"Long zipcode", "123456789"},
		{"Zipcode with letters", "1234567a"},
		{"Zipcode with spaces", "12345 67"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/v1/zipweather/"+tt.zipcode, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnprocessableEntity {
				t.Errorf("Expected status %d, got %d", http.StatusUnprocessableEntity, w.Code)
			}

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			if response["message"] != "invalid zipcode" {
				t.Errorf("Expected 'invalid zipcode' message, got %v", response["message"])
			}
		})
	}
}

// TestFetchTemperature_MissingAPIKey tests fetchTemperature without WEATHER_API_KEY
func TestFetchTemperature_MissingAPIKey(t *testing.T) {
	restoreEnv := setEnv("WEATHER_API_KEY", "")
	defer restoreEnv()

	_, err := fetchTemperature("São Paulo", "SP", "Brazil")
	if err == nil {
		t.Error("Expected error for missing WEATHER_API_KEY, got nil")
	}
	if !strings.Contains(err.Error(), "missing WEATHER_API_KEY") {
		t.Errorf("Expected 'missing WEATHER_API_KEY' error, got %v", err)
	}
}

// TestFetchTemperature_Success tests fetchTemperature with valid response
func TestFetchTemperature_Success(t *testing.T) {
	restoreEnv := setEnv("WEATHER_API_KEY", "test-key")
	defer restoreEnv()

	// Create a mock server that responds like WeatherAPI
	mockResp := WeatherResponse{}
	mockResp.Current.TempC = 22.5

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	// Mock http.Get to return the mock server response
	restoreHTTP := patchHTTPGet(func(url string) (*http.Response, error) {
		return http.Get(server.URL)
	})
	defer restoreHTTP()

	temp, err := fetchTemperature("São Paulo", "SP", "Brazil")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if temp != 22.5 {
		t.Errorf("Expected 22.5, got %v", temp)
	}
}

// TestFetchTemperature_Non200Status tests fetchTemperature with non-200 status
func TestFetchTemperature_Non200Status(t *testing.T) {
	restoreEnv := setEnv("WEATHER_API_KEY", "test-key")
	defer restoreEnv()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	defer server.Close()

	restoreHTTP := patchHTTPGet(func(url string) (*http.Response, error) {
		return http.Get(server.URL)
	})
	defer restoreHTTP()

	_, err := fetchTemperature("São Paulo", "SP", "Brazil")
	if err == nil {
		t.Error("Expected error for non-200 status, got nil")
	}
	if !strings.Contains(err.Error(), "weatherapi error") {
		t.Errorf("Expected 'weatherapi error', got %v", err)
	}
}

// TestFetchTemperature_InvalidJSON tests fetchTemperature with invalid JSON response
func TestFetchTemperature_InvalidJSON(t *testing.T) {
	restoreEnv := setEnv("WEATHER_API_KEY", "test-key")
	defer restoreEnv()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "invalid json")
	}))
	defer server.Close()

	restoreHTTP := patchHTTPGet(func(url string) (*http.Response, error) {
		return http.Get(server.URL)
	})
	defer restoreHTTP()

	_, err := fetchTemperature("São Paulo", "SP", "Brazil")
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

// TestFetchLocationFromZipCode_Success tests fetchLocationFromZipCode with valid response
func TestFetchLocationFromZipCode_Success(t *testing.T) {
	mockResp := LocalidadeResponse{
		Localidade: "São Paulo",
		Estado:     "SP",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	restoreHTTP := patchHTTPGet(func(url string) (*http.Response, error) {
		return http.Get(server.URL)
	})
	defer restoreHTTP()

	location, err := fetchLocationFromZipCode("01310100")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if location.Localidade != "São Paulo" || location.Estado != "SP" {
		t.Errorf("Expected São Paulo/SP, got %v/%v", location.Localidade, location.Estado)
	}
}

// TestFetchLocationFromZipCode_EmptyLocalidade tests handling of empty localidade
func TestFetchLocationFromZipCode_EmptyLocalidade(t *testing.T) {
	mockResp := LocalidadeResponse{
		Localidade: "",
		Estado:     "SP",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	restoreHTTP := patchHTTPGet(func(url string) (*http.Response, error) {
		return http.Get(server.URL)
	})
	defer restoreHTTP()

	_, err := fetchLocationFromZipCode("01310100")
	if err == nil {
		t.Error("Expected error for empty localidade, got nil")
	}
}

// TestFetchLocationFromZipCode_EmptyEstado tests handling of empty estado
func TestFetchLocationFromZipCode_EmptyEstado(t *testing.T) {
	mockResp := LocalidadeResponse{
		Localidade: "São Paulo",
		Estado:     "",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	restoreHTTP := patchHTTPGet(func(url string) (*http.Response, error) {
		return http.Get(server.URL)
	})
	defer restoreHTTP()

	_, err := fetchLocationFromZipCode("01310100")
	if err == nil {
		t.Error("Expected error for empty estado, got nil")
	}
}

// Benchmark tests
func BenchmarkRound(b *testing.B) {
	for i := 0; i < b.N; i++ {
		round(25.567, 1)
	}
}

func BenchmarkIsValidString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isValidString("São Paulo")
	}
}
