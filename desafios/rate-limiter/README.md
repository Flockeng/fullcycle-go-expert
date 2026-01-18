# Rate Limiter em Go

Um rate limiter desenvolvido em Go que permite limitar o número de requisições por segundo com base em endereço IP ou token de acesso.

## 📚 Como Funciona

### Princípio de Operação

O rate limiter utiliza uma abordagem de **contador fixo por janela de tempo**. Para cada identificador (IP ou token), o sistema mantém um contador que rastreia o número de requisições dentro de uma janela de 1 segundo.

**Fluxo de Verificação:**

1. **Extração do Identificador**: O middleware extrai o token do header `API_KEY` ou identifica o IP da requisição.
2. **Verificação de Bloqueio**: Verifica se o identificador está atualmente bloqueado devido a um limite excedido anteriormente.
3. **Incremento do Contador**: Se não estiver bloqueado, incrementa o contador de requisições para aquele identificador.
4. **Verificação de Limite**: Compara o contador com o limite configurado.
5. **Ação**: 
   - Se dentro do limite: permite a requisição
   - Se excedido: bloqueia o identificador pelo tempo configurado e retorna HTTP 429

### Token Sobrescreve IP

Quando um token é fornecido via header `API_KEY`:
- O sistema **ignora** a verificação por IP
- Usa apenas o limite configurado para o token
- Se o token tiver um limite customizado no Redis, usa esse limite; caso contrário, usa o limite padrão

**Exemplo:**
- Limite por IP: 10 req/s
- Limite padrão por token: 100 req/s
- Requisição com token → usa 100 req/s (ignora 10 req/s do IP)

### Persistência no Redis

O rate limiter armazena as seguintes informações no Redis:

- **Contadores**: `ip:<IP>` e `token:<TOKEN>` - Contadores que expiram após 1 segundo
- **Bloqueios**: `block:ip:<IP>` e `block:token:<TOKEN>` - Marcas de bloqueio com TTL configurável
- **Limites Customizados**: `token_limit:<TOKEN>` - Limites personalizados por token (sem expiração)

### Janela de Tempo

O contador é resetado a cada segundo através da expiração automática no Redis. Isso significa que:
- As requisições são contadas dentro de janelas de 1 segundo
- Após 1 segundo sem requisições, o contador é resetado automaticamente
- O bloqueio (quando o limite é excedido) tem duração configurável e independente da janela de contagem

### Comportamento em Alta Concorrência

O sistema utiliza operações atômicas do Redis (INCR) para garantir consistência em ambientes de alta concorrência, evitando race conditions quando múltiplas requisições chegam simultaneamente.

## Funcionalidades

- ✅ Limitação por endereço IP
- ✅ Limitação por token de acesso (header `API_KEY`)
- ✅ Configuração de limite via token sobrescreve limite por IP
- ✅ Middleware HTTP injetável
- ✅ Configuração via variáveis de ambiente ou arquivo `.env`
- ✅ Tempo de bloqueio configurável
- ✅ Armazenamento no Redis
- ✅ Strategy pattern para fácil troca de mecanismo de persistência
- ✅ Lógica separada do middleware

## Configuração

O rate limiter pode ser configurado através de variáveis de ambiente ou arquivo `.env`:

| Variável | Descrição | Padrão |
|----------|-----------|--------|
| `RATE_LIMIT_IP` | Limite de requisições por segundo por IP | 10 |
| `RATE_LIMIT_IP_BLOCK_TIME` | Tempo de bloqueio do IP em segundos | 300 |
| `RATE_LIMIT_TOKEN_DEFAULT` | Limite padrão de requisições por segundo por token | 100 |
| `RATE_LIMIT_TOKEN_BLOCK_TIME` | Tempo de bloqueio do token em segundos | 300 |
| `REDIS_HOST` | Host do Redis | localhost |
| `REDIS_PORT` | Porta do Redis | 6379 |
| `REDIS_PASSWORD` | Senha do Redis (opcional) | "" |
| `REDIS_DB` | Número do banco de dados Redis | 0 |

## Instalação

1. Clone o repositório:
```bash
git clone <repo-url>
cd rate-limiter
```

2. Instale as dependências:
```bash
go mod download
```

3. Configure as variáveis de ambiente. Você pode criar um arquivo `.env` na raiz do projeto ou definir variáveis de ambiente:
```bash
cp .env.example .env
# Edite o .env conforme necessário
```

4. Inicie os serviços com Docker Compose:
```bash
# Inicia Redis e aplicação
docker-compose up -d
```

A aplicação estará disponível em `http://localhost:8080` e o Redis em `localhost:6379`.

### Resposta quando o limite é excedido

Quando o limite é excedido, o servidor retorna:

- **Código HTTP:** 429 (Too Many Requests)
- **Mensagem:** `{"error": "you have reached the maximum number of requests or actions allowed within a certain time frame"}`

## Arquitetura

O projeto segue uma arquitetura modular:

```
rate-limiter/
├── config/          # Configuração e carregamento de variáveis de ambiente
├── storage/         # Interface e implementações de storage (Redis)
├── limiter/         # Lógica do rate limiter (separada do middleware)
├── middleware/      # Middleware HTTP para integração com servidores web
└── cmd/server/      # Servidor de exemplo
```

### Strategy Pattern

O projeto utiliza strategy pattern para permitir fácil troca de mecanismos de persistência:

```go
type Storage interface {
    Increment(key string, expiration time.Duration) (int64, error)
    SetBlock(key string, duration time.Duration) error
    IsBlocked(key string) (bool, error)
    Reset(key string) error
}
```

## 🧪 Testes

Para informações detalhadas sobre como executar os testes, cobertura de testes e estratégia de testes do projeto, consulte o [Guia de Testes](TESTING.md).
