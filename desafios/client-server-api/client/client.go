package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type CotacaoResponse struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatalf("Erro ao criar a requisição, %v", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Erro na requisição, %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Fatalf("Erro: resposta HTTP %d - %s", res.StatusCode, http.StatusText(res.StatusCode))
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Erro ao ler resposta, %v", err)
	}

	var cotacao CotacaoResponse
	err = json.Unmarshal(body, &cotacao)
	if err != nil {
		log.Fatalf("Erro ao decodificar JSON, %v", err)
	}

	err = os.WriteFile("cotacao.txt", []byte(fmt.Sprintf("Dólar: %s\n", cotacao.Bid)), 0644)
	if err != nil {
		log.Fatalf("Erro ao escrever o arquivo cotacao.txt, %v", err)
	}
}
