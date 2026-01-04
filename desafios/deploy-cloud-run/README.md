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

## Arquitetura

- **Build Stage**: Golang 1.23 com compilação estática (CGO_ENABLED=0)
- **Runtime Stage**: Imagem `scratch` com apenas o binário e certificados CA
- **Tamanho**: Imagem otimizada e leve (apenas o essencial)

## Tecnologias

- [Gin Web Framework](https://github.com/gin-gonic/gin) - Framework HTTP
- [ViaCEP API](https://viacep.com.br/) - Consulta de CEP
- [WeatherAPI](https://www.weatherapi.com/) - Dados climáticos
