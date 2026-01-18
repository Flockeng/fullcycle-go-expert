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
