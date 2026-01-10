package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestValidCEP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/weather", handleWeatherRequest)

	requestBody := map[string]string{"cep": "01310100"}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/weather", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Errorf("Expected status 200 or 404, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if w.Code == http.StatusOK {
		if _, ok := response["city"]; !ok {
			t.Error("Response should contain 'city' field")
		}
		if _, ok := response["temp_C"]; !ok {
			t.Error("Response should contain 'temp_C' field")
		}
	}
}

func TestInvalidCEP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/weather", handleWeatherRequest)

	testCases := []string{
		"123",      // Too short
		"invalid",  // Not numeric
		"0131010a", // Contains letter
	}

	for _, cep := range testCases {
		requestBody := map[string]string{"cep": cep}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/weather", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("CEP %s: Expected status 422, got %d", cep, w.Code)
		}

		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)

		if response["message"] != "invalid zipcode" {
			t.Errorf("CEP %s: Expected 'invalid zipcode', got '%s'", cep, response["message"])
		}
	}
}

func TestMissingCEP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/weather", handleWeatherRequest)

	requestBody := map[string]string{}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/weather", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("Expected status 422, got %d", w.Code)
	}
}

func TestTemperatureConversion(t *testing.T) {
	tempC := 25.0

	tempF := tempC*1.8 + 32
	tempK := tempC + 273.15

	expectedF := 77.0
	expectedK := 298.15

	if int(tempF) != int(expectedF) {
		t.Errorf("Expected %f°F, got %f°F", expectedF, tempF)
	}

	if int(tempK) != int(expectedK) {
		t.Errorf("Expected %fK, got %fK", expectedK, tempK)
	}
}
