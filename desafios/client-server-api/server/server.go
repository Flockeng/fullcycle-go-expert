package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type CotacaoResponse struct {
	Bid string `json:"bid"`
}

type Exchange struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type ExchangeUSDBRL struct {
	USDBRL Exchange `json:"USDBRL"`
}

type Document struct {
	ID   int             `gorm:"primaryKey"`
	Data json.RawMessage `gorm:"type:json"`
	gorm.Model
}

var DB *gorm.DB

func main() {
	initDB()
	initServerMux()
}

func initServerMux() {
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", exchangeUSDBRLHandler)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}

func initDB() {
	var err error

	DB, err = gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err = DB.AutoMigrate(&Document{})
	if err != nil {
		panic(err)
	}
}

func insertDocument(exchangeUSDBRL *ExchangeUSDBRL) error {
	jsonBytes, err := json.Marshal(exchangeUSDBRL)
	if err != nil {
		log.Printf("Erro ao codificar Json: %v", err)
		return err
	}

	doc := Document{Data: json.RawMessage(jsonBytes)}

	ctxDB, cancelDB := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancelDB()

	err = DB.WithContext(ctxDB).Create(&doc).Error
	if err != nil {
		log.Printf("Erro ao tentar persistir o documento, %v", err)
		return err
	}

	return nil
}

func exchangeUSDBRLHandler(w http.ResponseWriter, r *http.Request) {
	exchangeUSDBRL, err := getExchangeUSDBRL()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(CotacaoResponse{exchangeUSDBRL.USDBRL.Bid})

	go insertDocument(exchangeUSDBRL)
}

func getExchangeUSDBRL() (*ExchangeUSDBRL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		log.Printf("Erro ao criar a requisição: %v", err)
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Erro na requisição: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Erro ao ler resposta: %v", err)
		return nil, err
	}

	var exchangeUSDBRL ExchangeUSDBRL
	err = json.Unmarshal(body, &exchangeUSDBRL)
	if err != nil {
		log.Printf("Erro ao decodificar JSON: %v", err)
		return nil, err
	}

	return &exchangeUSDBRL, nil
}
