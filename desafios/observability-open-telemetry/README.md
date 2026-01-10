# Observability com OpenTelemetry

Projeto de demonstração de observabilidade em Go utilizando OpenTelemetry e Zipkin para rastreamento distribuído de requisições entre microserviços.

## 📋 Visão Geral

Este projeto implementa dois serviços Go que trabalham em conjunto para fornecer informações de clima baseadas em um CEP (Código de Endereçamento Postal brasileiro). O projeto utiliza OpenTelemetry para instrumentação e Zipkin para visualização de traces distribuídos.

### Arquitetura

```
Cliente HTTP
    ↓
Service-A 
    ↓
Service-B 
    ↓
ViaCEP API + Weather API
    ↓
Zipkin - Visualização de Traces
```

## 🚀 Serviços

### Service-A (Gateway)

**Localização:** `cmd/service-a/main.go`

Serviço principal que:
- Recebe requisições HTTP POST com um CEP em formato JSON
- Valida o CEP (deve conter exatamente 8 dígitos numéricos)
- Encaminha a requisição para o Service-B
- Retorna a resposta com informações de temperatura

**Endpoint:**
```bash
POST /
Content-Type: application/json

{
  "cep": "01310100"
}
```

**Resposta:**
```json
{
  "city": "São Paulo",
  "temp_C": 25.5,
  "temp_F": 77.9,
  "temp_K": 298.65
}
```

**Porta:** 8080

**Variáveis de Ambiente:**
- `PORT` - Porta do serviço (padrão: 8080)
- `SERVICE_B_URL` - URL do Service-B (padrão: http://localhost:8081)
- `ZIPKIN_URL` - URL do Zipkin (padrão: http://zipkin:9411/api/v2/spans)

### Service-B (Weather Service)

**Localização:** `cmd/service-b/main.go`

Serviço responsável por:
- Consultar o CEP na ViaCEP API para obter cidade e estado
- Consultar a Weather API para obter temperatura
- Converter temperatura entre escalas (Celsius, Fahrenheit, Kelvin)
- Registrar traces de todas as operações

**Endpoint:**
```bash
POST /weather
Content-Type: application/json

{
  "cep": "01310100"
}
```

**Resposta:**
```json
{
  "city": "São Paulo",
  "temp_C": 25.5,
  "temp_F": 77.9,
  "temp_K": 298.65
}
```

**Porta:** 8081

**Variáveis de Ambiente:**
- `PORT_B` - Porta do serviço (padrão: 8081)
- `WEATHER_API_KEY` - Chave da API WeatherAPI.com (obrigatória)
- `ZIPKIN_URL` - URL do Zipkin (padrão: http://zipkin:9411/api/v2/spans)

## 🔍 OpenTelemetry e Zipkin

### O que é OpenTelemetry?

OpenTelemetry é um conjunto de APIs, SDKs e ferramentas para instrumentação de aplicações, coletando dados de telemetria (traces, métricas e logs). Permite observabilidade distribuída sem depender de um fornecedor específico.

### O que é Zipkin?

Zipkin é uma plataforma de rastreamento distribuído que ajuda a reunir dados de timing de microserviços. Fornece uma interface web para visualizar traces e dependências entre serviços.

### Instrumentação no Projeto

Ambos os serviços utilizam:

```go
import "go.opentelemetry.io/otel/exporters/zipkin"

// Inicializa o tracer provider
exporter, err := zipkin.New("http://zipkin:9411/api/v2/spans")
tp := trace.NewTracerProvider(
    trace.WithBatcher(exporter),
    trace.WithResource(res),
)
otel.SetTracerProvider(tp)

// Cria spans para rastreamento
ctx, span := tracer.Start(context.Background(), "operationName")
defer span.End()

// Adiciona eventos ao span
span.AddEvent("evento importante")
```

**Visualizar Traces:** Acesse http://localhost:9411 no navegador

## ✅ Testes

### Service-A Tests

**Arquivo:** `cmd/service-a/main_test.go`

Testes implementados:

1. **TestIsValidCEP** - Valida a função de verificação de CEP
   - CEPs válidos: "01310100", "29902555", "12345678"
   - CEPs inválidos: "123" (muito curto), "0131010a" (letra), "invalid", "" (vazio), etc.

2. **TestHandleCEPRequest_ValidCEP** - Testa requisição com CEP válido
   - Executa requisição POST com CEP válido
   - Verifica status HTTP e presença de campos na resposta

Executar testes:
```bash
cd cmd/service-a
go test -v
```

### Service-B Tests

**Arquivo:** `cmd/service-b/main_test.go`

Testes implementados:

1. **TestValidCEP** - Testa requisição com CEP válido
   - Verifica se o status é 200 (sucesso) ou 404 (CEP não encontrado)
   - Valida presença dos campos obrigatórios na resposta

2. **TestInvalidCEP** - Testa requisição com CEPs inválidos
   - Valida rejeição de CEPs com formato incorreto
   - Verifica retorno de status 422 (Unprocessable Entity)

Executar testes:
```bash
cd cmd/service-b
go test -v
```

### Executar Todos os Testes

```bash
go test ./...
```

## 🐳 Executando com Docker

### Pré-requisitos

- Docker e Docker Compose instalados
- Chave API do WeatherAPI.com (obter em https://www.weatherapi.com/)

### Passos para Rodar

1. **Clone ou acesse o diretório do projeto:**
```bash
cd desafios/observability-open-telemetry
```

2. **Configure as variáveis de ambiente:**
```bash
# Crie um arquivo .env na raiz do projeto
WEATHER_API_KEY=sua_chave_aqui
```

3. **Inicie os serviços com Docker Compose:**
```bash
docker-compose up --build -d
```

Isso iniciará:
- **Zipkin** na porta 9411
- **Service-B** na porta 8081
- **Service-A** na porta 8080

4. **Aguarde os serviços ficarem prontos** 

### Testando o Projeto

**Fazer uma requisição:**
```bash
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{"cep": "01310100"}'
```

**Resposta esperada:**
```json
{
  "city": "São Paulo",
  "temp_C": 25.5,
  "temp_F": 77.9,
  "temp_K": 298.65
}
```

**Visualizar Traces no Zipkin:**
- Acesse http://localhost:9411

Exemplo de informações que você verá no Zipkin para um trace distribuído:

```text
Service: service-a
Trace ID: 4bf92f3577b34da6a3ce929d0e0e4736
Span: handleCEPRequest
Timestamp: 2026-01-10T12:34:56.789Z
Duration: 45ms
Tags/Attributes:
  - request.cep: 01310100
  - http.method: POST
  - http.url: http://service-b:8081/weather
  - http.status_code: 200

Child Span: callServiceB -> service-b/handleWeatherRequest
  - service: service-b
  - span: handleWeatherRequest
  - tags:
      - request.cep: 01310100
      - external.api: weatherapi
      - http.status_code: 200

External calls (examples):
  - viacep: http GET https://viacep.com.br/ws/01310100/json
    - tags: http.method=GET, http.url=https://viacep.com.br/..., external.api=viacep
  - weatherapi: http GET https://api.weatherapi.com/v1/current.json?...
    - tags: http.method=GET, http.url=https://api.weatherapi.com/..., external.api=weatherapi
```

### Parar os Serviços

```bash
docker-compose down
```

## 🔨 Estrutura do Projeto

```
.
├── docker-compose.yml          # Orquestração de containers
├── Dockerfile.service-a        # Build de Service-A
├── Dockerfile.service-b        # Build de Service-B
├── go.mod                      # Dependências Go
├── go.sum                      # Checksums das dependências
├── cmd/
│   ├── service-a/
│   │   ├── main.go             # Código principal Service-A
│   │   └── main_test.go        # Testes Service-A
│   └── service-b/
│       ├── main.go             # Código principal Service-B
│       └── main_test.go        # Testes Service-B
└── README.md                   # Este arquivo
```

## 🏗️ Nota sobre build/arquitetura

Os Dockerfiles (`Dockerfile.service-a` e `Dockerfile.service-b`) utilizam atualmente `GOARCH=amd64` no comando de build para compilar binários para a arquitetura amd64. Se você precisa rodar o projeto em outra arquitetura (por exemplo `arm64` em Macs com Apple Silicon), pode alterar o valor de `GOARCH` diretamente nos Dockerfiles.


## 📦 Dependências Principais

```
github.com/gin-gonic/gin                    # Framework HTTP
go.opentelemetry.io/otel                    # OpenTelemetry SDK
go.opentelemetry.io/otel/exporters/zipkin   # Exportador Zipkin
go.opentelemetry.io/otel/sdk                # Core SDK
```

## 🌐 APIs Externas Utilizadas

1. **ViaCEP** - Consulta de endereço por CEP
   - Endpoint: https://viacep.com.br/ws/{cep}/json
   - Retorna: cidade e estado

2. **WeatherAPI.com** - Informações de clima
   - Endpoint: https://api.weatherapi.com/v1/current.json
   - Requer: API key válida

## 📖 Mais Informações

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Zipkin Documentation](https://zipkin.io/pages/quickstart)
- [Gin Documentation](https://gin-gonic.com/)
- [ViaCEP API](https://viacep.com.br/)
- [WeatherAPI.com](https://www.weatherapi.com/)
