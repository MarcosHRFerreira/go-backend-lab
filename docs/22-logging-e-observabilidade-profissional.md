# Estudo De Logging E Observabilidade Profissional No Go Tweets

## Objetivo

Este material descreve como evoluir o `go-tweets` para um padrao profissional e atual de observabilidade, com foco em:

- logs estruturados
- rastreabilidade por requisicao
- correlacao entre HTTP, service e repository
- visibilidade de erros internos sem vazar detalhes ao cliente
- metricas operacionais
- trilha para tracing distribuido no futuro

O objetivo nao e apenas "colocar logs", mas construir um modelo de observabilidade que permita responder perguntas reais de producao:

- Qual rota mais falha?
- Quanto tempo cada endpoint demora?
- Em qual camada o erro aconteceu?
- O problema foi validacao, autenticacao, banco ou bug interno?
- Qual foi a sequencia de eventos de uma requisicao especifica?
- O banco esta lento?
- O refresh token falhou por regra de negocio ou erro tecnico?

---

## Diagnostico Do Estado Atual

Hoje o projeto ja tem uma boa base estrutural para evoluir observabilidade:

- bootstrap centralizado em [main.go](file:///c:/Users/marco/OneDrive/Projects/go-tweets/cmd/main.go)
- tratamento HTTP padronizado em [response.go](file:///c:/Users/marco/OneDrive/Projects/go-tweets/internal/httpresponse/response.go)
- erro de aplicacao com `statusCode` e `cause` em [error.go](file:///c:/Users/marco/OneDrive/Projects/go-tweets/internal/apperror/error.go)
- middlewares centralizados em [middleware.go](file:///c:/Users/marco/OneDrive/Projects/go-tweets/internal/middleware/middleware.go)
- fluxo em camadas bem definido: handler -> service -> repository

Mesmo assim, a observabilidade ainda esta em nivel basico.

### O Que Existe Hoje

1. Logs pontuais de ciclo de vida no bootstrap:
   - inicio do servidor
   - sinal de desligamento
   - encerramento gracioso
   - falha no fechamento do banco

2. Access log basico do Gin:
   - `gin.Logger()`
   - util para desenvolvimento
   - insuficiente para producao madura

3. Recovery do Gin:
   - `gin.Recovery()`
   - protege contra panic
   - mas nao cria um padrao completo de observabilidade

4. Contrato HTTP de erro:
   - o cliente recebe mensagens limpas
   - porem o sistema nao registra contexto interno suficiente para investigacao

### Lacunas Principais

1. Nao existe logging estruturado em JSON.
2. Nao existe `request_id` para correlacionar uma requisicao ponta a ponta.
3. Nao existe logger central injetado nas camadas.
4. Nao existe padrao para campos de log.
5. Nao existe distincao clara entre:
   - erro esperado de negocio
   - erro de dependencia
   - falha de seguranca
   - bug interno
6. Nao existem metricas HTTP, metricas de banco ou contadores de erro.
7. Nao existe tracing com contexto propagado.
8. Nao existe trilha de auditoria para eventos sensiveis como login, refresh token e acoes autenticadas.

---

## Visao Profissional Atual

Em um backend Go moderno, o padrao profissional de observabilidade normalmente combina:

1. `log/slog` para logging estruturado
2. middleware de `request_id`
3. middleware de access log customizado
4. middleware de recovery com log estruturado
5. OpenTelemetry para traces e metricas
6. Prometheus para scraping de metricas
7. Grafana para dashboards
8. eventualmente Loki, Elasticsearch ou Datadog para consulta de logs

Para o seu projeto, a estrategia mais equilibrada e:

1. Fase 1: `slog` + request correlation + access log customizado
2. Fase 2: erros estruturados + logs por camada
3. Fase 3: metricas HTTP e banco
4. Fase 4: tracing com OpenTelemetry

Essa ordem e importante porque evita complexidade excessiva cedo demais.

---

## Recomendacao Tecnica Para O Go Tweets

### Stack Recomendada

- Logger principal: `log/slog`
- Formato em runtime: JSON
- Correlacao: `request_id`
- Metricas: Prometheus
- Tracing: OpenTelemetry
- Dashboard: Grafana

### Por Que `slog`

Para um projeto Go moderno e didatico, `slog` e a escolha mais alinhada com o mercado atual porque:

- faz parte da biblioteca padrao
- oferece logs estruturados
- reduz dependencia externa
- integra bem com contexto
- e suficiente para a maior parte dos backends

Se voce estivesse construindo uma plataforma de altissimo throughput com muita customizacao de logging, `zap` tambem seria uma boa opcao. Mas para este projeto, `slog` entrega melhor equilibrio entre padrao moderno, simplicidade e valor educacional.

---

## Arquitetura Alvo

### 1. Logger Central

Criar um pacote interno, por exemplo:

```text
internal/observability/logger
```

Esse pacote deve:

- criar o `slog.Logger`
- configurar formato JSON
- definir nivel de log
- permitir enriquecer logs com `service`, `env` e `version`

Exemplo conceitual:

```go
package logger

import (
	"log/slog"
	"os"
)

func New() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	return slog.New(handler).With(
		slog.String("service", "go-tweets"),
	)
}
```

### 2. Injetar Logger No Bootstrap

Em [main.go](file:///c:/Users/marco/OneDrive/Projects/go-tweets/cmd/main.go), o logger deve nascer logo no inicio e ser usado em todo o ciclo de vida:

- config carregada
- conexao com banco aberta
- servidor iniciado
- shutdown iniciado
- shutdown finalizado

### 3. Middleware De Request ID

Cada requisicao deve receber um identificador unico:

- se o cliente enviar `X-Request-ID`, reaproveita
- se nao enviar, o servidor gera

Esse valor deve entrar em:

- headers de resposta
- contexto do request
- todos os logs da requisicao

### 4. Access Log Estruturado

Substituir `gin.Logger()` por middleware proprio para registrar:

- metodo
- rota
- path
- status code
- latencia
- ip
- user agent
- request_id
- user_id quando existir

Exemplo de campos:

```json
{
  "level": "INFO",
  "msg": "http request completed",
  "request_id": "9f1c3b0f",
  "method": "POST",
  "route": "/auth/login",
  "status_code": 200,
  "latency_ms": 18,
  "client_ip": "127.0.0.1"
}
```

### 5. Recovery Com Log Estruturado

`gin.Recovery()` deve dar lugar a um recovery proprio ou ser complementado com middleware que:

- captura panic
- registra stack trace
- inclui `request_id`
- retorna `500` limpo ao cliente

### 6. Logger No Contexto

O logger enriquecido da requisicao deve ser recuperavel por contexto.

Exemplo de atributos adicionados por request:

- `request_id`
- `method`
- `path`
- `user_id` quando autenticado

Assim, service e repository conseguem registrar eventos sem recriar metadados manualmente.

---

## Onde Logar No Projeto

Um erro comum e logar em toda linha. Isso gera ruido. O profissional e logar nos pontos certos.

### Camada `main`

Logar:

- inicio da aplicacao
- configuracao carregada
- conexao com banco pronta
- servidor ouvindo porta
- sinal de desligamento
- falha de shutdown

Nao logar:

- detalhes sensiveis de env
- senha, DSN completo ou segredo JWT

### Camada `middleware`

Logar:

- requisicao iniciada
- requisicao concluida
- falha de autenticacao
- panic recuperado

Campos importantes:

- `request_id`
- `method`
- `path`
- `status_code`
- `latency_ms`
- `user_id`

### Camada `handler`

Handler normalmente nao deve logar tudo.

Logar apenas quando:

- houver evento de borda importante
- houver erro relevante antes de chegar ao service
- a requisicao for sensivel em termos de seguranca

Exemplos:

- tentativa de login
- validacao de payload malformado
- falha de autorizacao

### Camada `service`

Aqui esta o melhor lugar para logar eventos de negocio.

Exemplos no seu projeto:

- `Register`: usuario criado
- `Login`: autenticacao bem-sucedida
- `RefreshToken`: token rotacionado
- `CreatePost`: post criado
- `UpdatePost`: post atualizado
- `DeletePost`: post excluido logicamente
- `LikeOrUnlikePost`: like inserido ou removido
- `CreateComment`: comentario criado

Importante: service nao deve logar o mesmo erro varias vezes se o handler ja vai registrar a falha final. O ideal e definir responsabilidade clara.

### Camada `repository`

Repository deve logar com moderacao.

Logar:

- falhas de banco
- timeout de query
- operacoes lentas

Evitar:

- logar cada query em nivel `INFO`
- logar payloads completos
- duplicar log do mesmo erro em varias camadas

### Camada `httpresponse`

Hoje [response.go](file:///c:/Users/marco/OneDrive/Projects/go-tweets/internal/httpresponse/response.go) retorna respostas limpas, o que esta certo.

O proximo passo profissional e:

- continuar escondendo detalhes do cliente
- mas registrar internamente o `cause` do `apperror`

Ou seja:

- cliente recebe: `failed to create token`
- log interno registra: erro tecnico detalhado, stack ou causa encadeada

---

## Taxonomia De Logs Recomendada

### `DEBUG`

Usar apenas em desenvolvimento.

Exemplos:

- valores intermediarios
- decisao interna de toggle
- detalhes de mapeamento

### `INFO`

Usar para eventos normais importantes.

Exemplos:

- server started
- login succeeded
- refresh token rotated
- post created

### `WARN`

Usar para comportamento anormal, mas esperado.

Exemplos:

- token invalido
- tentativa de acessar recurso inexistente
- payload invalido
- ownership mismatch escondido como `not found`

### `ERROR`

Usar quando houve falha tecnica real.

Exemplos:

- falha ao abrir banco
- falha ao executar query
- falha ao gerar JWT
- panic recuperado

---

## Campos Padrao De Log

Todo log profissional precisa de padrao. Sugestao para o projeto:

### Campos Fixos De Plataforma

- `service`
- `env`
- `version`
- `component`

### Campos De Requisicao

- `request_id`
- `trace_id` no futuro
- `method`
- `path`
- `route`
- `status_code`
- `latency_ms`
- `client_ip`
- `user_agent`

### Campos De Identidade

- `user_id`
- `username` apenas se fizer sentido e sem exagero

### Campos De Banco

- `db_operation`
- `table`
- `duration_ms`
- `rows_affected`

### Campos De Erro

- `error`
- `cause`
- `error_kind`
- `stack` apenas em panic ou debug controlado

---

## Eventos Criticos Do Seu Projeto

### Autenticacao

Logar:

- tentativa de login
- login bem-sucedido
- falha de credenciais
- refresh token invalido
- refresh token rotacionado

Nao logar:

- senha
- refresh token completo
- JWT completo

Se precisar identificar token em investigacao, registre:

- prefixo mascarado
- hash do token
- ultimos 6 caracteres

### Cadastro

Logar:

- tentativa de registro
- conflito de email ou username
- usuario criado

### Posts

Logar:

- criacao
- atualizacao
- exclusao logica
- like e unlike

### Comentarios

Logar:

- criacao
- like e unlike

---

## Relacao Entre Logging, Erros E Auditoria

E importante separar tres coisas:

### 1. Log Operacional

Serve para diagnosticar problema tecnico.

Exemplo:

- timeout no banco
- panic
- latencia alta

### 2. Log De Negocio

Serve para entender comportamento do sistema.

Exemplo:

- usuario criou post
- usuario curtiu comentario

### 3. Log De Auditoria

Serve para trilha de seguranca e conformidade.

Exemplo:

- login bem-sucedido
- token renovado
- tentativa de uso de token invalido
- acao autenticada de exclusao

No `go-tweets`, autenticacao e alteracoes de conteudo merecem ao menos um nivel basico de auditoria.

---

## Plano De Implementacao Recomendado

## Etapa 1 - Estruturar O Logger

Criar:

```text
internal/observability/logger/logger.go
```

Responsabilidades:

- construir `slog.Logger`
- definir formato JSON
- nivel por ambiente

## Etapa 2 - Request ID

Criar:

```text
internal/middleware/request_id.go
```

Responsabilidades:

- ler `X-Request-ID`
- gerar novo ID se necessario
- propagar para contexto e resposta

## Etapa 3 - Access Log Customizado

Criar:

```text
internal/middleware/access_log.go
```

Responsabilidades:

- medir tempo da requisicao
- recuperar status final
- enriquecer log com usuario autenticado

Substituir em [main.go](file:///c:/Users/marco/OneDrive/Projects/go-tweets/cmd/main.go):

- remover `gin.Logger()`
- adicionar middleware proprio

## Etapa 4 - Recovery Estruturado

Criar:

```text
internal/middleware/recovery.go
```

Responsabilidades:

- capturar panic
- logar stack trace
- devolver `500`

## Etapa 5 - Logger No Contexto

Criar helper, por exemplo:

```text
internal/observability/logctx/logctx.go
```

Responsabilidades:

- guardar logger no contexto
- recuperar logger enriquecido

## Etapa 6 - Logar Eventos De Negocio

Aplicar em services prioritarios:

- [register.go](file:///c:/Users/marco/OneDrive/Projects/go-tweets/internal/service/user/register.go)
- [login.go](file:///c:/Users/marco/OneDrive/Projects/go-tweets/internal/service/user/login.go)
- [refresh_token.go](file:///c:/Users/marco/OneDrive/Projects/go-tweets/internal/service/user/refresh_token.go)
- [create_post.go](file:///c:/Users/marco/OneDrive/Projects/go-tweets/internal/service/post/create_post.go)
- [update_post.go](file:///c:/Users/marco/OneDrive/Projects/go-tweets/internal/service/post/update_post.go)
- [delete_post.go](file:///c:/Users/marco/OneDrive/Projects/go-tweets/internal/service/post/delete_post.go)
- [create_comment.go](file:///c:/Users/marco/OneDrive/Projects/go-tweets/internal/service/comment/create_comment.go)

## Etapa 7 - Metricas

Adicionar:

- contador de requests por rota
- contador de erros por status
- histograma de latencia HTTP
- metrica de duracao de query

Sugestao:

- endpoint `/metrics`
- Prometheus scrape
- dashboards no Grafana

## Etapa 8 - Tracing

Depois da base pronta, integrar OpenTelemetry para:

- `trace_id`
- spans por request
- spans por banco
- correlacao com logs

---

## Exemplo De Politica De Logging Por Fluxo

### `POST /auth/login`

1. Middleware:
   - request started
   - request completed

2. Service:
   - login attempted
   - login failed por credencial invalida em `WARN`
   - login succeeded em `INFO`

3. Caso erro tecnico:
   - `ERROR` com `cause`

### `POST /auth/refresh-token`

1. Middleware:
   - request started
   - request completed

2. Service:
   - refresh token requested
   - active user reloaded
   - previous refresh token removed
   - new refresh token stored
   - token rotation succeeded

### `POST /tweets`

1. Middleware:
   - request started
   - request completed

2. Service:
   - create post requested
   - post created

### `DELETE /tweets/:post_id`

1. Middleware:
   - request started
   - request completed

2. Service:
   - delete requested
   - ownership mismatch em `WARN`
   - post soft deleted em `INFO`

---

## O Que Nao Deve Ser Logado

Nunca registrar em texto puro:

- senha
- hash de senha completo
- JWT completo
- refresh token completo
- segredo JWT
- DSN completo do banco
- conteudo sensivel de `.env`

Tambem evitar:

- corpo completo de request de autenticacao
- resposta completa de erro interno
- dados pessoais sem necessidade operacional

---

## Anti-Patterns Comuns

### 1. Logar O Mesmo Erro Em Todas As Camadas

Resultado:

- ruido
- duplicidade
- dificuldade para consulta

Melhor:

- logar onde faz mais sentido diagnostico
- responder ao cliente em outro ponto

### 2. Log Texto Solto

Exemplo ruim:

```text
deu erro ao salvar
```

Melhor:

```json
{
  "level": "ERROR",
  "msg": "failed to store refresh token",
  "component": "user_service",
  "request_id": "ab12cd34",
  "user_id": 7,
  "cause": "duplicate key"
}
```

### 3. Logar Tudo Em `ERROR`

Credencial invalida nao e erro tecnico. E evento esperado.

Melhor:

- `WARN` para tentativa invalida
- `ERROR` para falha real de infraestrutura ou bug

### 4. Misturar Logging Com Regra De Negocio

Service nao deve ficar poluido por mensagens irrelevantes.

Melhor:

- poucos logs
- logs com intencao
- logs nos pontos de decisao

---

## Mapa Aplicado Ao Projeto

### Arquivos Mais Importantes Para A Primeira Implementacao

- [main.go](file:///c:/Users/marco/OneDrive/Projects/go-tweets/cmd/main.go)
- [middleware.go](file:///c:/Users/marco/OneDrive/Projects/go-tweets/internal/middleware/middleware.go)
- [response.go](file:///c:/Users/marco/OneDrive/Projects/go-tweets/internal/httpresponse/response.go)
- [error.go](file:///c:/Users/marco/OneDrive/Projects/go-tweets/internal/apperror/error.go)
- [mysql.go](file:///c:/Users/marco/OneDrive/Projects/go-tweets/pkg/internalsql/mysql.go)

### Pontos De Ganho Rapido

1. substituir `log` por `slog` em [main.go](file:///c:/Users/marco/OneDrive/Projects/go-tweets/cmd/main.go)
2. remover `gin.Logger()` e criar access log proprio
3. adicionar `request_id`
4. logar `cause` de `apperror.Internal(...)`
5. adicionar auditoria minima em login, refresh token e delete post

---

## Roadmap Recomendado Para Voce

### Sprint 1

- logger central com `slog`
- request id
- access log estruturado
- recovery estruturado

### Sprint 2

- logger no contexto
- logs nos services de auth, post e comment
- padrao de campos

### Sprint 3

- metricas Prometheus
- dashboards basicos
- alertas simples

### Sprint 4

- OpenTelemetry
- tracing HTTP + banco
- correlacao `trace_id` com logs

---

## Conclusao

Hoje o `go-tweets` esta organizado o suficiente para receber observabilidade de forma limpa, mas ainda opera com logging basico.

O caminho profissional atual para esse projeto e:

1. adotar `slog`
2. estruturar logs em JSON
3. criar correlacao por `request_id`
4. registrar eventos certos nas camadas certas
5. adicionar metricas
6. evoluir para tracing

Se essa estrategia for aplicada, o projeto deixa de ter apenas "mensagens no console" e passa a ter uma base de observabilidade real, adequada para estudo serio e mais proxima do que o mercado espera em APIs backend modernas.

---

## Proximo Passo Sugerido

O proximo passo mais valioso e implementar a Fase 1 inteira no codigo:

- `slog`
- request id
- access log customizado
- recovery estruturado

Depois disso, fica muito mais natural expandir para metricas, auditoria e tracing.
