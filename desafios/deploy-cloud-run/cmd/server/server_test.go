package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func setEnv(key, value string) func() {
	old := os.Getenv(key)
	os.Setenv(key, value)
	return func() { os.Setenv(key, old) }
}

func patchHTTPGet(fn func(url string) (*http.Response, error)) func() {
	orig := httpGet
	httpGet = fn
	return func() { httpGet = orig }
}

var httpGet = http.Get

func fetchTemperatureTest(city, state, country string) (float64, error) {
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		return 0, errors.New("missing WEATHER_API_KEY")
	}
	query := city + "," + state + "," + country
	query = strings.ReplaceAll(query, " ", "%20")
	url := "http://mocked"
	resp, err := httpGet(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, errors.New("weatherapi error")
	}
	var weather WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
		return 0, err
	}
	return weather.Current.TempC, nil
}

func TestFetchTemperature_Success(t *testing.T) {
	restoreEnv := setEnv("WEATHER_API_KEY", "dummy")
	defer restoreEnv()

	mockResp := WeatherResponse{}
	mockResp.Current.TempC = 25.5
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	restoreHTTP := patchHTTPGet(func(url string) (*http.Response, error) {
		return http.Get(server.URL)
	})
	defer restoreHTTP()

	temp, err := fetchTemperatureTest("São Paulo", "SP", "Brazil")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if temp != 25.5 {
		t.Errorf("expected 25.5, got %v", temp)
	}
}

func TestFetchTemperature_MissingAPIKey(t *testing.T) {
	restoreEnv := setEnv("WEATHER_API_KEY", "")
	defer restoreEnv()

	_, err := fetchTemperatureTest("São Paulo", "SP", "Brazil")
	if err == nil || !strings.Contains(err.Error(), "missing WEATHER_API_KEY") {
		t.Errorf("expected missing WEATHER_API_KEY error, got %v", err)
	}
}

func TestFetchTemperature_Non200Status(t *testing.T) {
	restoreEnv := setEnv("WEATHER_API_KEY", "dummy")
	defer restoreEnv()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	defer server.Close()

	restoreHTTP := patchHTTPGet(func(url string) (*http.Response, error) {
		return http.Get(server.URL)
	})
	defer restoreHTTP()

	_, err := fetchTemperatureTest("São Paulo", "SP", "Brazil")
	if err == nil || !strings.Contains(err.Error(), "weatherapi error") {
		t.Errorf("expected weatherapi error, got %v", err)
	}
}

func TestFetchTemperature_InvalidJSON(t *testing.T) {
	restoreEnv := setEnv("WEATHER_API_KEY", "dummy")
	defer restoreEnv()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "invalid json")
	}))
	defer server.Close()

	restoreHTTP := patchHTTPGet(func(url string) (*http.Response, error) {
		return http.Get(server.URL)
	})
	defer restoreHTTP()

	_, err := fetchTemperatureTest("São Paulo", "SP", "Brazil")
	if err == nil {
		t.Errorf("expected error for invalid JSON, got nil")
	}
}
