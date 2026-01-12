# 🎯 Sistema de Leilões (Auction System)

Sistema de leilões desenvolvido em Go que permite criar leilões, fazer lances e gerenciar usuários. O sistema possui fechamento automático de leilões após um intervalo configurável.

## 📋 Pré-requisitos

- Docker
- Docker Compose

## 🚀 Como Executar o Projeto

### 1. Configurar Variáveis de Ambiente

Configure o arquivo `.env` no diretório `cmd/auction/`, segue abaixo exemplo de um arquivo configurado:

```env
MONGODB_URL=mongodb://admin:admin@mongodb:27017/?authSource=admin
MONGODB_DB=auctions
AUCTION_INTERVAL=5m
MONGO_INITDB_ROOT_USERNAME=admin
MONGO_INITDB_ROOT_PASSWORD=admin
```

**Variáveis de Ambiente:**
- `MONGODB_URL`: URL de conexão com o MongoDB
- `MONGODB_DB`: Nome do banco de dados
- `AUCTION_INTERVAL`: Intervalo para fechamento automático dos leilões (ex: `5m`, `10m`, `1h`)
- `MONGO_INITDB_ROOT_USERNAME`: Usuário root do MongoDB
- `MONGO_INITDB_ROOT_PASSWORD`: Senha root do MongoDB

### 2. Executar com Docker Compose

Execute o seguinte comando na raiz do projeto:

```bash
docker-compose up --build
```

Este comando irá:
- Construir a imagem da aplicação Go
- Iniciar o container do MongoDB
- Iniciar o container da aplicação na porta `8080`

### 3. Verificar se está Funcionando

Acesse `http://localhost:8080` ou teste os endpoints da API. A aplicação estará rodando na porta `8080`.

### 4. Parar os Containers

Para parar os containers:

```bash
docker-compose down
```

Para parar e remover os volumes (dados do MongoDB):

```bash
docker-compose down -v
```

## 📡 Endpoints da API

- `GET /auction` - Listar todos os leilões
- `GET /auction/:auctionId` - Buscar leilão por ID
- `POST /auction` - Criar novo leilão
- `GET /auction/winner/:auctionId` - Buscar lance vencedor do leilão
- `POST /bid` - Criar novo lance
- `GET /bid/:auctionId` - Buscar lances de um leilão
- `GET /user` - Listar todos os usuários
- `GET /user/:userId` - Buscar usuário por ID

## 📝 Exemplos de Uso com CURL

### Criar um Leilão

```bash
curl -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "Sony Alpha ZV",
    "category": "Electronics",
    "description": "Sony Alpha ZV in perfect condition, sealed box with original accessories",
    "condition": 1
  }'
```

**Resposta esperada:**
```json
{
  "id": "a8f062c1-572e-43a0-9b4f-669034f817fa",
  "product_name": "Sony Alpha ZV",
  "category": "Electronics",
  "description": "Sony Alpha ZV in perfect condition, sealed box with original accessories",
  "condition": 1,
  "status": 0,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Buscar Leilão por ID

```bash
curl http://localhost:8080/auction/a8f062c1-572e-43a0-9b4f-669034f817fa
```

**Resposta esperada:**
```json
{
  "id": "a8f062c1-572e-43a0-9b4f-669034f817fa",
  "product_name": "Sony Alpha ZV",
  "category": "Electronics",
  "description": "Sony Alpha ZV in perfect condition, sealed box with original accessories",
  "condition": 1,
  "status": 0,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Criar um Lance (Bid)

```bash
curl -X POST http://localhost:8080/bid \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "a8f062c1-572e-43a0-9b4f-669034f817fa",
    "auction_id": "a8f062c1-572e-43a0-9b4f-669034f817fa",
    "amount": 1500.00
  }'
```

**Resposta esperada:**
```json
{
  "id": "b9f173d2-683f-54b1-ac5g-770145g928gb",
  "user_id": "a8f062c1-572e-43a0-9b4f-669034f817fa",
  "auction_id": "a8f062c1-572e-43a0-9b4f-669034f817fa",
  "amount": 1500.00,
  "timestamp": "2024-01-15T10:35:00Z"
}
```

### Buscar Lance Vencedor (Winning Bid)

```bash
curl http://localhost:8080/auction/winner/a8f062c1-572e-43a0-9b4f-669034f817fa
```

**Resposta esperada:**
```json
{
  "id": "b9f173d2-683f-54b1-ac5g-770145g928gb",
  "user_id": "a8f062c1-572e-43a0-9b4f-669034f817fa",
  "auction_id": "a8f062c1-572e-43a0-9b4f-669034f817fa",
  "amount": 1500.00,
  "timestamp": "2024-01-15T10:35:00Z"
}
```

**Nota:** Os IDs nos exemplos acima são apenas ilustrativos. Use os IDs retornados pelas respostas da API em suas requisições subsequentes.

## 🧪 Testes

### Teste de Fechamento Automático de Leilões

O arquivo `internal/infra/database/auction/create_auction_test.go` contém testes que validam o fechamento automático dos leilões.

#### O que o teste faz?

O teste `TestCreateAuction_AutomaticClosure` valida que:

1. **Criação do Leilão**: Cria um leilão com status `Active`
2. **Verificação Inicial**: Confirma que o leilão foi criado com status `Active`
3. **Aguardar Intervalo**: Aguarda o intervalo configurado (2 segundos no teste) + buffer de segurança
4. **Verificação Final**: Confirma que o leilão foi automaticamente atualizado para status `Completed`

O teste valida que a goroutine em `CreateAuction` está funcionando corretamente, atualizando o status do leilão após o intervalo definido pela variável de ambiente `AUCTION_INTERVAL`.

#### Como Executar o Teste

**Pré-requisito**: Certifique-se de que o MongoDB está rodando (pode ser via `docker-compose up mongodb` ou uma instância local).

Execute o teste com o comando:

```bash
go test -v ./internal/infra/database/auction -run TestCreateAuction_AutomaticClosure
```

**Executar todos os testes do pacote:**

```bash
go test -v ./internal/infra/database/auction
```

**Executar com cobertura:**

```bash
go test -v -cover ./internal/infra/database/auction -run TestCreateAuction_AutomaticClosure
```

#### Detalhes do Teste

- **Intervalo configurado**: O teste define `AUCTION_INTERVAL=2s` para acelerar a execução
- **Tempo de espera**: O teste aguarda 3 segundos (intervalo + buffer)
- **Limpeza**: O teste limpa a coleção antes e depois da execução
- **Conexão MongoDB**: Usa variáveis de ambiente ou valores padrão se não configuradas

#### Exemplo de Saída Esperada

```
=== RUN   TestCreateAuction_AutomaticClosure
    create_auction_test.go:78: Aguardando fechamento automático do leilão (intervalo: 2s)...
    create_auction_test.go:89: Fechamento automático funcionou corretamente! Status atualizado para Completed.
--- PASS: TestCreateAuction_AutomaticClosure (3.02s)
PASS
ok  	fullcycle-auction_go/internal/infra/database/auction	5.037s
```

## 🔧 Tecnologias Utilizadas

- **Go 1.20**: Linguagem de programação
- **Gin**: Framework web
- **MongoDB**: Banco de dados NoSQL
- **Docker**: Containerização
- **Docker Compose**: Orquestração de containers

