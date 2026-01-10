package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var tracer = otel.Tracer("service-b")

type LocationResponse struct {
	Localidade string `json:"localidade"`
	Estado     string `json:"estado"`
}

type WeatherResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

type TemperatureResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

func initTracer() (*trace.TracerProvider, error) {
	zipkinURL := os.Getenv("ZIPKIN_URL")
	if zipkinURL == "" {
		zipkinURL = "http://zipkin:9411/api/v2/spans"
	}

	exporter, err := zipkin.New(zipkinURL)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("service-b"),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	return tp, nil
}

func main() {
	tp, err := initTracer()
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalf("Failed to shutdown tracer: %v", err)
		}
	}()

	router := gin.Default()
	router.POST("/weather", handleWeatherRequest)

	port := os.Getenv("PORT_B")
	if port == "" {
		port = "8081"
	}

	log.Printf("Service B started on port %s", port)
	router.Run(":" + port)
}

func handleWeatherRequest(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "handleWeatherRequest")
	defer span.End()

	var requestBody map[string]string
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "invalid zipcode"})
		return
	}

	zipcode, exists := requestBody["cep"]
	if !exists {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "invalid zipcode"})
		return
	}

	// Add request CEP and HTTP information to the span
	span.SetAttributes(
		attribute.String("request.cep", zipcode),
		attribute.String("http.method", "POST"),
		attribute.String("http.route", "/weather"),
	)

	matched, _ := regexp.MatchString(`^\d{8}$`, zipcode)
	if !matched {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "invalid zipcode"})
		return
	}

	location, err := fetchLocationFromZipCode(ctx, zipcode)
	if err != nil {
		span.RecordError(err)
		log.Printf("fetchLocationFromZipCode error: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"message": "can not find zipcode"})
		return
	}

	tempC, err := fetchTemperature(ctx, location.Localidade, location.Estado, "Brazil")
	if err != nil {
		span.RecordError(err)
		log.Printf("fetchTemperature error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
		return
	}

	tempF := tempC*1.8 + 32
	tempK := tempC + 273.15

	response := TemperatureResponse{
		City:  location.Localidade,
		TempC: round(tempC, 1),
		TempF: round(tempF, 1),
		TempK: round(tempK, 1),
	}

	c.JSON(http.StatusOK, response)
}

func fetchLocationFromZipCode(ctx context.Context, zipcode string) (*LocationResponse, error) {
	_, span := tracer.Start(ctx, "fetchLocationFromZipCode")
	defer span.End()
	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json", zipcode)

	// Create request with context and inject propagation headers
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	// Add attributes about external call
	span.SetAttributes(
		attribute.String("http.method", "GET"),
		attribute.String("http.url", url),
		attribute.String("external.api", "viacep"),
	)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	var locationResponse LocationResponse
	if err := json.NewDecoder(resp.Body).Decode(&locationResponse); err != nil {
		return nil, err
	}

	if locationResponse.Localidade == "" || locationResponse.Estado == "" {
		return nil, fmt.Errorf("empty location or state")
	}

	return &locationResponse, nil
}

func fetchTemperature(ctx context.Context, city string, state string, country string) (float64, error) {
	_, span := tracer.Start(ctx, "fetchTemperature")
	defer span.End()

	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		return 0, fmt.Errorf("missing WEATHER_API_KEY")
	}

	query := city + "," + state + "," + country
	query = strings.ReplaceAll(query, " ", "%20")
	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s", apiKey, query)

	// Create request with context and inject propagation headers
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	// Add attributes about weather API call
	span.SetAttributes(
		attribute.String("http.method", "GET"),
		attribute.String("http.url", url),
		attribute.String("external.api", "weatherapi"),
	)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var weatherResponse WeatherResponse
	if err := json.Unmarshal(body, &weatherResponse); err != nil {
		return 0, err
	}

	return weatherResponse.Current.TempC, nil
}

func round(value float64, precision int) float64 {
	multiplier := 1.0
	for i := 0; i < precision; i++ {
		multiplier *= 10
	}
	return float64(int(value*multiplier)) / multiplier
}
