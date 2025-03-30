package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Response struct {
	Body string
	Err  error
}

func process(ch chan Response, url string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		ch <- Response{Err: err}
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			ch <- Response{Err: errors.New("Timeout!")}
		} else {
			ch <- Response{Err: err}
		}
		return
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		msgError := "API is responding in an invalid format, Content-Type:" + contentType
		ch <- Response{Err: errors.New(msgError)}
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ch <- Response{Err: err}
		return
	}

	json := Response{Body: string(body)}
	err = validateJSON(json.Body)
	if err != nil {
		ch <- Response{Err: err}
		return
	}
	ch <- json
}

func validateJSON(content string) error {
	var jsonMap map[string]any
	err := json.Unmarshal([]byte(content), &jsonMap)
	if err != nil {
		return err
	}

	//brasilapi - Exemplo de mensagem de erro para CEP 12345678 invalido
	// {"message":"Todos os serviços de CEP retornaram erro.",
	// "type":"service_error","name":"CepPromiseError",
	// "errors":[
	// 	{"name":"ServiceError","message":"A autenticacao de null falhou!","service":"correios"},
	// 	{"name":"ServiceError","message":"Cannot read properties of undefined (reading 'replace')","service":"viacep"},
	// 	{"name":"ServiceError","message":"Erro ao se conectar com o serviço WideNet.","service":"widenet"},
	// 	{"name":"ServiceError","message":"CEP não encontrado na base dos Correios.","service":"correios-alt"}]}
	_, result := jsonMap["errors"]
	if result {
		return errors.New(content)
	}

	//viacep - Exemplo de mensagem de erro para CEP 12345678 invalido
	// {"erro":"true"}
	_, result = jsonMap["erro"]
	if result {
		return errors.New(content)
	}

	return nil
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Error: Invalid argument, please enter the zip code when running the program. Example: go run main.go 01153000")
		os.Exit(1)
	}

	cep := os.Args[1]

	if len(cep) != 8 {
		fmt.Println("Error: The number must be exactly 8 digits long.")
		os.Exit(1)
	}

	_, err := strconv.Atoi(cep)
	if err != nil {
		fmt.Println("Error: The argument must be a valid integer.")
		os.Exit(1)
	}

	urlBrasilApi := "https://brasilapi.com.br/api/cep/v1/" + cep
	urlViaCep := "http://viacep.com.br/ws/" + cep + "/json/"

	chanViaCep := make(chan Response)
	chanBrasilApi := make(chan Response)

	go process(chanViaCep, urlViaCep)
	go process(chanBrasilApi, urlBrasilApi)

	select {
	case msg := <-chanViaCep:
		if msg.Err != nil {
			fmt.Println("Received from viacep, ERROR:", msg.Err)
		} else {
			fmt.Println("Received from viacep, JSON:", msg.Body)
		}

	case msg := <-chanBrasilApi:
		if msg.Err != nil {
			fmt.Println("Received from brasilapi, ERROR:", msg.Err)
		} else {
			fmt.Println("Received from brasilapi, JSON:", msg.Body)
		}
	case <-time.After(time.Second):
		fmt.Println("Timeout!")
	}

}
