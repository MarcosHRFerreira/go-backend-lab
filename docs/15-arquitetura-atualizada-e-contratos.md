# Arquitetura Atualizada e Contratos da API

## Objetivo

Este documento registra o estado atual do projeto apos a rodada de modernizacao feita no codigo.

Ele complementa os capitulos anteriores e deve ser lido como a referencia mais atual para:

- bootstrap da aplicacao
- contrato de erro HTTP
- validacao de entrada
- responsabilidades entre `handler`, `service` e `repository`
- nomenclaturas principais que foram ajustadas

## O Que Mudou em Relacao a Versao Inicial

Nas primeiras versoes do projeto, havia um acoplamento maior entre regra de negocio e HTTP.
Em especial:

- `service` retornava `statusCode` junto com `error`
- validacoes e mensagens estavam espalhadas
- bootstrap do servidor era funcional, mas menos robusto
- a documentacao refletia esse momento anterior

Hoje o projeto esta mais proximo de um backend Go de mercado para estudo guiado:

- a camada HTTP ficou mais padronizada
- o bootstrap ficou mais seguro
- o tratamento de erro ficou mais consistente
- o projeto ganhou testes unitarios e de integracao

## Nova Estrutura Conceitual

O fluxo principal agora fica assim:

```text
HTTP -> Gin Router -> Middleware -> Handler -> Service -> Repository -> MySQL
```

Mas com uma diferenca importante em relacao ao desenho anterior:

- o `handler` traduz HTTP
- o `service` decide regra de negocio
- o `repository` cuida de persistencia
- o erro de aplicacao trafega como `error`, sem `statusCode` espalhado

## Bootstrap Mais Profissional

## `cmd/main.go`

O `main.go` agora faz mais do que apenas subir o Gin e chamar `Run`.

Pontos importantes:

- usa `gin.New()` em vez de `gin.Default()`
- adiciona `Logger` e `Recovery` explicitamente
- cria `http.Server` com timeouts
- executa `db.PingContext` no health check
- faz shutdown gracioso com `os/signal`
- fecha o banco no encerramento

### Timeouts configurados

- `ReadHeaderTimeout`
- `ReadTimeout`
- `WriteTimeout`
- `IdleTimeout`
- timeout de shutdown

Isso torna o bootstrap mais realista para ambiente de producao e excelente para estudo de operacao backend.

## `internal/config/config.go`

O carregamento de configuracao tambem amadureceu:

- `.env` passou a ser opcional
- `PORT` tem default `8080`
- `DB_PORT` tem default `3306`
- campos obrigatorios sao validados no boot

Campos exigidos:

- `SECRET_JWT`
- `DB_HOST`
- `DB_USER`
- `DB_NAME`
- `DB_PASSWORD`

### Ganho didatico

Isso ensina um principio importante de backend:

- falhar cedo quando configuracao obrigatoria nao existe

## `pkg/internalsql/mysql.go`

A conexao MySQL tambem foi fortalecida:

- funcao principal passou a ser `ConnectMySQL`
- a funcao antiga `ConectMySql` ficou como compatibilidade
- a DSN agora inclui `charset`, `collation`, `timeout`, `readTimeout` e `writeTimeout`
- o codigo faz `PingContext` antes de seguir
- o pool de conexoes e configurado explicitamente

Configuracoes principais do pool:

- `SetMaxOpenConns`
- `SetMaxIdleConns`
- `SetConnMaxLifetime`
- `SetConnMaxIdleTime`

## Contrato de Erro de Aplicacao

## `internal/apperror/error.go`

Foi criado um pacote especifico para erro de aplicacao.

O objetivo foi remover este antigo padrao:

```go
resultado, statusCode, err := service.Algo(...)
```

E migrar para algo mais idiomatico:

```go
resultado, err := service.Algo(...)
```

Com isso:

- o `service` nao precisa mais devolver status HTTP cru
- o `handler` continua responsavel por traduzir resposta HTTP
- o erro carrega semantica de API sem misturar demais a camada

### Tipos principais

- `apperror.New(statusCode, message)`
- `apperror.BadRequest(message)`
- `apperror.Unauthorized(message)`
- `apperror.NotFound(message)`
- `apperror.Internal(message, cause)`

### Ideia central

O `error` continua sendo um erro Go normal, mas agora tambem sabe informar o `statusCode`.

## Helpers HTTP Compartilhados

## `internal/httpresponse/response.go`

Foi criado um ponto central para respostas HTTP padronizadas.

Principais helpers:

- `JSONError(...)`
- `JSONAppError(...)`
- `AbortJSONError(...)`
- `BindAndValidateJSON(...)`
- `ParseIntParam(...)`
- `ParseMinInt64Query(...)`

### O que isso resolveu

- reduziu duplicacao de bind e validate nos handlers
- padronizou mensagens de erro
- centralizou resposta de validacao
- padronizou parse de parametros e query params

## Formato atual de erro

Erro simples:

```json
{
  "message": "invalid authorization token"
}
```

Erro de validacao:

```json
{
  "message": "validation failed",
  "errors": [
    {
      "field": "email",
      "message": "must be a valid email"
    }
  ]
}
```

## Middleware de Autenticacao

## `internal/middleware/middleware.go`

O middleware foi alinhado ao novo contrato de resposta.

Melhorias visiveis:

- `missing token` virou `missing authorization token`
- token invalido virou `invalid authorization token`
- respostas usam JSON consistente
- funcao principal passou a usar `jwt.ValidateToken(...)`

### Diferenca entre os middlewares

- `AuthMiddleware(secretKey)` valida claims normalmente
- `AuthRefreshTokenMiddleware(secretKey)` usa parse sem validacao de claims para o fluxo de refresh

Esse segundo caso continua sendo um ponto de discussao arquitetural importante para estudo de seguranca.

## Separacao de Responsabilidades Atual

## Handler

Responsavel por:

- bind de JSON
- leitura de path params e query params
- validacao de entrada
- resposta HTTP final

Ele nao deveria conter regra de negocio mais pesada.

## Service

Responsavel por:

- regras de negocio
- coordenacao entre repositories
- decisao de cenarios alternativos
- devolucao de erro de aplicacao

Exemplos atuais:

- `Login` devolve `Unauthorized` para credencial invalida
- `RefreshToken` devolve `Unauthorized` quando o refresh token nao bate
- `DetailPost` devolve `NotFound` para post ausente

## Repository

Responsavel por:

- SQL manual
- `ExecContext`, `QueryContext`, `QueryRowContext`
- mapeamento de resultados
- tratamento de `sql.ErrNoRows`

## Ajustes de Nomenclatura Ja Feitos

Alguns nomes visiveis foram alinhados:

- `ValidadeToken` ganhou a forma correta `ValidateToken`
- `LikeOrUnlikecomment` virou `LikeOrUnlikeComment`

Ainda existe espaco para evoluir naming no restante do projeto, mas os pontos mais visiveis da trilha principal ja ficaram melhores para estudo.

## O Que Ainda Nao E O Estado Final

Mesmo com a modernizacao, o projeto ainda nao representa o topo absoluto de maturidade de mercado.
Ele esta em um ponto muito melhor para estudo realista, mas ainda pode evoluir em:

- logging estruturado
- observabilidade
- mensagens de sucesso mais consistentes
- estrategia mais forte para refresh token
- revisao de naming em arquivos e metodos menos centrais

## Como Recomendo Ler o Projeto Agora

Se voce quiser estudar o estado atual e nao apenas o historico do projeto, siga esta ordem:

1. `00-overview.md`
2. `01-bootstrap-config.md`
3. `03-auth-and-utils.md`
4. este arquivo `15-arquitetura-atualizada-e-contratos.md`
5. `10-main-e-auth-line-by-line.md`
6. `11-get-all-post-line-by-line.md`
7. `16-testes-automatizados.md`

## Conclusao

Este projeto deixou de ser apenas um CRUD inicial para virar uma boa base de estudo sobre:

- bootstrap em Go
- contrato HTTP
- erro de aplicacao
- middleware com Gin
- SQL manual
- separacao de camadas
- refatoracao incremental orientada a boas praticas

Por isso, ao estudar os capitulos antigos, use este documento como complemento oficial da arquitetura atual.
