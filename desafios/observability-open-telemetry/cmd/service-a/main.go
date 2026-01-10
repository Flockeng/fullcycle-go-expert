package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var tracer = otel.Tracer("service-a")

type CEPRequest struct {
	CEP string `json:"cep" binding:"required"`
}

type WeatherResponse struct {
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
			semconv.ServiceNameKey.String("service-a"),
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
	router.POST("/", handleCEPRequest)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Service A started on port %s", port)
	router.Run(":" + port)
}

func handleCEPRequest(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "handleCEPRequest")
	defer span.End()

	var req CEPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.AddEvent("invalid request")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "invalid zipcode"})
		return
	}

	// Validate CEP format (must be 8 digits)
	if !isValidCEP(req.CEP) {
		span.AddEvent("invalid CEP format")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "invalid zipcode"})
		return
	}

	// Add CEP as span attribute (useful for filtering). Avoid adding overly sensitive data in production.
	span.SetAttributes(attribute.String("request.cep", req.CEP))

	// Call Service B
	response, statusCode, err := callServiceB(ctx, req.CEP)
	if err != nil {
		c.JSON(statusCode, response)
		return
	}

	c.JSON(statusCode, response)
}

func isValidCEP(cep string) bool {
	if len(cep) != 8 {
		return false
	}
	for _, char := range cep {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func callServiceB(ctx context.Context, cep string) (interface{}, int, error) {
	_, span := tracer.Start(ctx, "callServiceB")
	defer span.End()

	serviceBURL := os.Getenv("SERVICE_B_URL")
	if serviceBURL == "" {
		serviceBURL = "http://service-b:8081"
	}

	requestBody := map[string]string{"cep": cep}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return gin.H{"message": "internal error"}, http.StatusInternalServerError, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", serviceBURL+"/weather", bytes.NewBuffer(jsonBody))
	if err != nil {
		return gin.H{"message": "internal error"}, http.StatusInternalServerError, err
	}

	req.Header.Set("Content-Type", "application/json")

	// Inject trace context into outgoing request headers so Zipkin can correlate spans.
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	// Add attributes to the span about the outgoing HTTP request
	span.SetAttributes(
		attribute.String("http.method", "POST"),
		attribute.String("http.url", serviceBURL+"/weather"),
		attribute.String("http.target", "/weather"),
		attribute.String("request.cep", cep),
	)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return gin.H{"message": "internal error"}, http.StatusInternalServerError, err
	}
	defer resp.Body.Close()

	// Record HTTP response status as span attribute
	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	var responseBody interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return gin.H{"message": "internal error"}, http.StatusInternalServerError, err
	}

	return responseBody, resp.StatusCode, nil
}
