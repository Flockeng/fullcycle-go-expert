package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type LocalidadeResponse struct {
	Localidade string `json:"localidade"`
	Estado     string `json:"estado"`
}

type WeatherResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

var httpGet = http.Get

func main() {
	router := gin.Default()
	router.GET("/v1/zipweather/:zipcode", getWeatherByZipCode)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run(":" + port)
}

func getWeatherByZipCode(c *gin.Context) {
	zipcode := c.Param("zipcode")

	matched, _ := regexp.MatchString(`^\d{8}$`, zipcode)
	if !matched {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "invalid zipcode"})
		return
	}

	cep, err := fetchLocationFromZipCode(zipcode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "can not find zipcode"})
		return
	}

	tempC, err := fetchTemperature(cep.Localidade, cep.Estado, "Brazil")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
		return
	}

	tempF := tempC*1.8 + 32
	tempK := tempC + 273

	c.JSON(http.StatusOK, gin.H{
		"temp_C": round(tempC, 1),
		"temp_F": round(tempF, 1),
		"temp_K": round(tempK, 1),
	})
}

func fetchLocationFromZipCode(paramZipCode string) (*LocalidadeResponse, error) {
	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json", paramZipCode)
	resp, err := httpGet(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var localidadeResponse LocalidadeResponse
	if err := json.NewDecoder(resp.Body).Decode(&localidadeResponse); err != nil {
		return nil, err
	}

	if !isValidString(localidadeResponse.Localidade) {
		return nil, fmt.Errorf("localidade is empty")
	}

	if !isValidString(localidadeResponse.Estado) {
		return nil, fmt.Errorf("estado is empty")
	}

	return &localidadeResponse, nil
}

func fetchTemperature(city string, state string, country string) (float64, error) {
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		return 0, fmt.Errorf("missing WEATHER_API_KEY")
	}

	query := city + "," + state + "," + country
	query = strings.ReplaceAll(query, " ", "%20")

	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s&lang=pt", apiKey, query)

	resp, err := httpGet(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("weatherapi error")
	}

	var weather WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
		return 0, err
	}

	return weather.Current.TempC, nil
}

func isValidString(value string) bool {
	return strings.TrimSpace(value) != ""
}

func round(val float64, precision int) float64 {
	format := "%." + strconv.Itoa(precision) + "f"
	v, _ := strconv.ParseFloat(fmt.Sprintf(format, val), 64)
	return v
}
