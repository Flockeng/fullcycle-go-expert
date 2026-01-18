# Guia de Testes - Rate Limiter

Este documento explica como os testes foram estruturados e como executá-los.

## Estrutura de Testes

### Testes Unitários

Os testes unitários não dependem de serviços externos e utilizam um storage em memória (`MemoryStorage`) para simular o comportamento do Redis.

#### `storage/storage_test.go`
Testa a implementação em memória do storage:
- Incremento de contadores
- Bloqueio e verificação de bloqueio
- Reset de chaves
- Limites customizados de tokens

#### `limiter/limiter_test.go`
Testa a lógica do rate limiter:
- Limitação por IP
- Limitação por token
- Limites customizados por token
- Independência entre diferentes IPs/tokens

#### `middleware/middleware_test.go`
Testa o middleware HTTP:
- Limitação por IP através do middleware
- Limitação por token através do middleware
- Token sobrescrevendo limite de IP
- Extração de token de diferentes formatos de header
- Respostas HTTP 429 corretas

### Testes de Integração

Os testes de integração (`integration/redis_integration_test.go`) testam a aplicação com Redis real:
- Conexão e operações no Redis
- Rate limiting end-to-end com Redis
- Limites customizados de tokens no Redis

## Executando os Testes

### Todos os Testes Unitários

```bash
go test ./...
```

**Nota:** Os testes de integração não são executados por padrão porque utilizam build tags (`//go:build integration`). Eles só rodam quando você usa explicitamente `-tags=integration`.

### Testes de Integração

**Pré-requisito:** Redis deve estar rodando.

```bash
go test ./integration -tags=integration -v
```

### Todos os Testes (Unitários + Integração)

```bash
go test ./... -tags=integration -v
```
## Casos de Teste Cobertos

### Limitação por IP
- ✅ Permite requisições até o limite configurado
- ✅ Bloqueia após exceder o limite
- ✅ Mantém bloqueio pelo tempo configurado
- ✅ IPs diferentes têm limites independentes

### Limitação por Token
- ✅ Permite requisições até o limite do token
- ✅ Bloqueia após exceder o limite
- ✅ Usa limite padrão quando token não tem limite customizado
- ✅ Usa limite customizado quando configurado
- ✅ Tokens diferentes têm limites independentes

### Comportamento do Middleware
- ✅ Extrai token do header `API_KEY`
- ✅ Extrai token do header `Authorization: API_KEY <token>`
- ✅ Retorna HTTP 429 quando limite é excedido
- ✅ Retorna mensagem de erro correta
- ✅ Token sobrescreve verificação por IP
- ✅ Quando não há token, verifica apenas por IP

### Integração com Redis
- ✅ Persistência de contadores
- ✅ Persistência de bloqueios
- ✅ Limites customizados de tokens
- ✅ Operações atômicas em alta concorrência

## Testes de Carga Manual

Para testar o comportamento sob carga, use ferramentas externas:

### Apache Bench
```bash
# 1000 requisições, 10 conexões simultâneas

# Sem token
ab -n 1000 -c 10 http://localhost:8080/

# Com token
ab -n 1000 -c 10 -H "API_KEY: meu-token" http://localhost:8080/
```


#### Resultado sem token
Benchmarking localhost (be patient)
Completed 100 requests
Completed 200 requests
Completed 300 requests
Completed 400 requests
Completed 500 requests
Completed 600 requests
Completed 700 requests
Completed 800 requests
Completed 900 requests
Completed 1000 requests
Finished 1000 requests


Server Software:        
Server Hostname:        localhost
Server Port:            8080

Document Path:          /
Document Length:        33 bytes

Concurrency Level:      10
Time taken for tests:   0.122 seconds
Complete requests:      1000
Failed requests:        990
   (Connect: 0, Receive: 0, Length: 990, Exceptions: 0)
Non-2xx responses:      990
Total transferred:      230100 bytes
HTML transferred:       106260 bytes
Requests per second:    8205.33 [#/sec] (mean)
Time per request:       1.219 [ms] (mean)
Time per request:       0.122 [ms] (mean, across all concurrent requests)
Transfer rate:          1843.80 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   0.1      0       0
Processing:     0    1   1.0      1      10
Waiting:        0    1   1.0      1      10
Total:          0    1   1.0      1      11

Percentage of the requests served within a certain time (ms)
  50%      1
  66%      1
  75%      1
  80%      1
  90%      2
  95%      2
  98%      3
  99%     10
 100%     11 (longest request)

#### Resultado com token
Benchmarking localhost (be patient)
Completed 100 requests
Completed 200 requests
Completed 300 requests
Completed 400 requests
Completed 500 requests
Completed 600 requests
Completed 700 requests
Completed 800 requests
Completed 900 requests
Completed 1000 requests
Finished 1000 requests

Server Software:        
Server Hostname:        localhost
Server Port:            8080

Document Path:          /
Document Length:        33 bytes

Concurrency Level:      10
Time taken for tests:   0.124 seconds
Complete requests:      1000
Failed requests:        900
   (Connect: 0, Receive: 0, Length: 900, Exceptions: 0)
Non-2xx responses:      900
Total transferred:      222000 bytes
HTML transferred:       99600 bytes
Requests per second:    8045.70 [#/sec] (mean)
Time per request:       1.243 [ms] (mean)
Time per request:       0.124 [ms] (mean, across all concurrent requests)
Transfer rate:          1744.28 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   0.1      0       1
Processing:     0    1   0.5      1       5
Waiting:        0    1   0.5      1       5
Total:          1    1   0.6      1       5

Percentage of the requests served within a certain time (ms)
  50%      1
  66%      1
  75%      1
  80%      1
  90%      2
  95%      2
  98%      3
  99%      4
 100%      5 (longest request)