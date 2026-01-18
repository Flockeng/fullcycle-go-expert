# Stress Tester (CLI em Go) ⚡️

Aplicação CLI para realizar testes de carga simples em um serviço web.

## Parâmetros via CLI:

- `--url` : URL do serviço a ser testado (obrigatório)
- `--requests` : Número total de requests a serem executados (obrigatório)
- `--concurrency` : Número de chamadas simultâneas (obrigatório)

## Exemplo de execução local:

```
go build -o stresstester
./stresstester --url=https://example.com --requests=1000 --concurrency=10
```

## Exemplo usando Docker:

```
docker build -t stresstester:latest .
docker run --rm stresstester:latest --url=https://example.com --requests=1000 --concurrency=10
```

### O relatório exibido ao final inclui:

- Tempo total gasto na execução
- Quantidade total de requests realizados
- Quantidade de requests com status HTTP 200
- Distribuição de outros códigos de status HTTP (404, 500, etc.)
- Total de erros de conexão/falha

## Exemplo usando http429.com que é um serviço de teste de limite de taxa

```
docker run --rm stresstester:latest --url=https://http429.com/ratelimit/10/5 --requests=1000 --concurrency=10
```

### Resultado do relatório:

```
--- Relatório ---
Tempo total gasto: 34.62770864s
Quantidade total de requests realizados: 1000
Quantidade de requests com status HTTP 200: 338
Distribuição de códigos de status HTTP:
  200: 338
  429: 662
--- Fim do relatório ---
```