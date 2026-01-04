# Weather Server

Servidor Go que retorna a temperatura atual de uma localidade brasileira a partir de um CEP.

## Sobre o Projeto

Este projeto implementa uma API HTTP que:
- Recebe um CEP (8 dígitos) como parâmetro
- Consulta a API ViaCEP para obter localidade e estado
- Consulta a API WeatherAPI para obter a temperatura atual
- Retorna a temperatura em Celsius, Fahrenheit e Kelvin

### Endpoint

```
GET /v1/zipweather/:zipcode
```

**Exemplo:**
```bash
curl http://localhost:8080/v1/zipweather/01310100
```

**Resposta de sucesso (200):**
```json
{
  "temp_C": 25.5,
  "temp_F": 77.9,
  "temp_K": 298.65
}
```

**Erro - CEP inválido (422):**
```json
{
  "message": "invalid zipcode"
}
```

**Erro - CEP não encontrado (404):**
```json
{
  "message": "can not find zipcode"
}
```

## Requisitos

- Go 1.23+ (para execução local)
- Docker (para containerização)
- Chave de API do [WeatherAPI](https://www.weatherapi.com/)

## Como Rodar

### Via Docker Compose (Recomendado)

1. Configure a variável de ambiente com sua chave de API:
```bash
export WEATHER_API_KEY=sua_chave_aqui
```

2. Execute o serviço:
```bash
docker-compose up --build
```

3. A API estará disponível em `http://localhost:8080`

### Via Docker (sem Compose)

```bash
docker build -t zipweather .
docker run -e WEATHER_API_KEY=sua_chave_aqui -p 8080:8080 zipweather
```

### Localmente (sem Docker)

1. Configure a variável de ambiente:
```bash
export WEATHER_API_KEY=sua_chave_aqui
```

2. Execute o servidor:
```bash
cd cmd/server
go run server.go
```

3. A API estará disponível em `http://localhost:8080`

## Configuração

### Variáveis de Ambiente

- `WEATHER_API_KEY` (obrigatório): Chave de API do WeatherAPI

## Testes

O projeto inclui testes unitários abrangentes em [cmd/server/server_test.go](cmd/server/server_test.go).

### Executar Testes

```bash
cd cmd/server
go test -v
```

### Testes com Cobertura

```bash
cd cmd/server
go test -v -cover
```

Para gerar relatório detalhado de cobertura:
```bash
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Cobertura de Testes

Os testes cobrem os seguintes cenários:

**Helper Functions:**
- `TestIsValidString` — Valida strings vazias, com espaços e válidas
- `TestRound` — Testa arredondamento com diferentes precisões
- `BenchmarkRound` — Benchmark da função de arredondamento
- `BenchmarkIsValidString` — Benchmark da validação de string

**API Endpoint:**
- `TestGetWeatherByZipCode_InvalidZipcode` — Rejeita CEP com formato inválido (< 8, > 8, com letras, com espaços)
- Retorna status 422 com mensagem "invalid zipcode"

**fetchTemperature:**
- `TestFetchTemperature_MissingAPIKey` — Valida erro quando `WEATHER_API_KEY` não está definida
- `TestFetchTemperature_Success` — Testa requisição bem-sucedida com mock server
- `TestFetchTemperature_Non200Status` — Testa tratamento de erro HTTP não-200
- `TestFetchTemperature_InvalidJSON` — Testa tratamento de resposta JSON inválida

**fetchLocationFromZipCode:**
- `TestFetchLocationFromZipCode_Success` — Testa sucesso na consulta de CEP
- `TestFetchLocationFromZipCode_EmptyLocalidade` — Rejeita localidade vazia
- `TestFetchLocationFromZipCode_EmptyEstado` — Rejeita estado vazio

### Mocking

Os testes usam mocking para HTTP requests através da variável global `httpGet` que pode ser substituída por uma função mock usando `patchHTTPGet()`.

## Arquitetura

- **Build Stage**: Golang 1.23 com compilação estática (CGO_ENABLED=0)
- **Runtime Stage**: Imagem `scratch` com apenas o binário e certificados CA
- **Tamanho**: Imagem otimizada e leve (apenas o essencial)

## Tecnologias

- [Gin Web Framework](https://github.com/gin-gonic/gin) - Framework HTTP
- [ViaCEP API](https://viacep.com.br/) - Consulta de CEP
- [WeatherAPI](https://www.weatherapi.com/) - Dados climáticos

## Nota sobre Google Cloud Run

**⚠️ Deploy não disponibilizado via Google Cloud Run**

O projeto **não está deployado no Google Cloud Run** por razões de viabilidade financeira. O modelo de preços do Google Cloud Run:

- **Antes**: Free tier verdadeiramente gratuito
- **Agora**: Requer **prepagamento mínimo de R$ 200** mesmo para usar o free tier “teste gratuito”, sendo creditado como saldo normal. (04/01/2026)

Para fazer deploy do serviço no Google Cloud Run, seria necessário:
1. Criar uma conta Google Cloud
2. Configurar conta de faturamento
3. Efetuar o deploy via `gcloud run deploy`

Se deseja explorar essa opção, consulte a [documentação do Google Cloud Run](https://cloud.google.com/run/docs).