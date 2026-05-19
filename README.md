# Go Tweets

API backend em Go para estudo de arquitetura, autenticacao, persistencia relacional, testes automatizados e boas praticas de organizacao de projeto.

Este repositorio foi evoluido como um projeto de estudo avancado e portfolio tecnico, com foco em:

- arquitetura em camadas
- tratamento de erros padronizado
- autenticacao com JWT e refresh token
- integracao com MySQL
- testes unitarios e de integracao
- documentacao tecnica detalhada

## Objetivo

O objetivo deste projeto e servir como um laboratorio pratico de backend em Go para quem quer sair do nivel de sintaxe e estudar uma aplicacao mais proxima de cenarios reais.

Ele foi usado para consolidar aprendizado em:

- construcao de APIs HTTP com Gin
- separacao entre handler, service e repository
- validacao de entrada e contratos HTTP
- uso de `context`, timeouts e graceful shutdown
- SQL manual com `database/sql`
- testes automatizados
- documentacao tecnica e trilha de estudo

## Stack

- Go
- Gin
- MySQL
- JWT
- `database/sql`
- `go-playground/validator`
- testes com `testing`, `httptest` e stubs
- `golangci-lint`

## Principais Funcionalidades

- cadastro de usuario
- login
- refresh token
- criacao de tweets
- listagem paginada de tweets
- detalhamento de tweet
- atualizacao e exclusao logica de tweet
- like e unlike de tweet
- criacao de comentarios
- like e unlike de comentario
- health check da aplicacao e do banco

## Arquitetura

O projeto segue uma organizacao em camadas para deixar responsabilidades mais claras:

- `cmd`: bootstrap da aplicacao e subida do servidor HTTP
- `internal/config`: carregamento e validacao de configuracoes
- `internal/handler`: camada HTTP
- `internal/service`: regras de negocio
- `internal/repository`: acesso a dados
- `internal/dto`: contratos de entrada e saida
- `internal/model`: estruturas de dominio usadas no projeto
- `internal/middleware`: autenticacao e middleware HTTP
- `internal/httpresponse`: respostas padronizadas e parsing de request
- `internal/apperror`: padronizacao de erros de aplicacao
- `pkg/internalsql`: conexao com banco, JWT e refresh token
- `test/unit`: testes unitarios
- `test/integration`: testes de integracao HTTP

## Fluxo da Requisicao

De forma resumida, o fluxo principal da API segue este caminho:

`Request HTTP -> middleware -> handler -> DTO -> service -> repository -> MySQL -> response HTTP`

Esse modelo ajuda a separar:

- transporte HTTP
- validacao
- regra de negocio
- persistencia

## Endpoints Principais

### Auth

- `POST /auth/register`
- `POST /auth/login`
- `POST /auth/refresh`

### Tweets

- `GET /tweets/`
- `GET /tweets/:post_id/detail`
- `POST /tweets/`
- `PUT /tweets/:post_id/update`
- `DELETE /tweets/:post_id/delete`
- `POST /tweets/action`

### Comments

- `POST /comment/`
- `POST /comment/action`

### Health

- `GET /check-health`

## Melhorias Tecnicas Ja Aplicadas

Ao longo da evolucao do projeto, foram aplicadas melhorias importantes de corretude e maturidade:

- contrato de erro HTTP mais consistente
- validacao centralizada de request
- parse e validacao de paginacao
- tratamento de erros ignorados em operacoes SQL
- bootstrap com configuracao validada
- conexao com banco com `PingContext`
- timeouts de servidor
- graceful shutdown
- middleware de autenticacao mais robusto
- testes unitarios de services
- testes de integracao dos principais endpoints

## Como Rodar Localmente

### 1. Suba o banco com Docker

```bash
docker compose up -d
```

O `docker-compose.yml` sobe um MySQL 8 com configuracao local para estudo.

### 2. Configure as variaveis de ambiente

O projeto le estas variaveis:

- `PORT`
- `SECRET_JWT`
- `DB_HOST`
- `DB_USER`
- `DB_NAME`
- `DB_PASSWORD`
- `DB_PORT`
- `DATABASE_URL`

Exemplo de `.env`:

```env
PORT=8080
SECRET_JWT=uma-chave-segura
DB_HOST=127.0.0.1
DB_USER=dbeaver
DB_NAME=go_tweets
DB_PASSWORD=superSecret
DB_PORT=3306
DATABASE_URL=mysql://dbeaver:superSecret@tcp(127.0.0.1:3306)/go_tweets
```

### 3. Execute as migrations

As migrations ficam em `db/migrations`.

Se estiver usando `dbmate`, rode algo como:

```bash
npx dbmate up
```

### 4. Inicie a API

```bash
go run ./cmd
```

Por padrao, a aplicacao sobe em `http://localhost:8080`.

## Testes

Para rodar toda a suite:

```bash
go test ./...
```

Tipos de teste presentes no projeto:

- unitarios em `test/unit`
- integracao HTTP em `test/integration`

## Qualidade de Codigo

Ferramentas e praticas usadas no projeto:

- `gofmt`
- `go test ./...`
- `go vet ./...`
- `golangci-lint run`

## Documentacao de Estudo

Este projeto possui uma trilha de documentacao essencial pensada para explicar a arquitetura e os fluxos principais do sistema.

Arquivos importantes:

- [docs/README.md](docs/README.md)

O material inclui:

- explicacao da arquitetura
- fluxo principal da aplicacao
- organizacao das camadas
- schema e migrations
- estrategia de testes automatizados

Materiais de estudo pessoais, apostilas impressas, guias consolidados e planos de carreira podem ser mantidos localmente, mas nao fazem parte da versao essencial do repositorio.

## Posicionamento do Projeto

Este repositorio deve ser lido como:

- projeto de estudo avancado
- projeto de portfolio
- laboratorio de backend em Go

Ele nao tenta se vender como sistema corporativo em producao, mas como uma base tecnica forte para demonstrar aprendizado, organizacao, evolucao e capacidade de entrega em Go.

## Proximos Passos Possiveis

- adicionar CI com GitHub Actions
- incluir colecao de requests ou arquivo `.http`
- documentar exemplos de payload por endpoint
- adicionar observabilidade mais profunda
- incluir benchmarks em fluxos criticos

## Autor

Projeto utilizado como base de estudo e evolucao tecnica por `Marcos H R Ferreira`, com foco na migracao de Java para Go e preparacao para vagas backend em Go.
