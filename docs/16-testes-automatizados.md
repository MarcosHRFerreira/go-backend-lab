# Testes Automatizados no Projeto

## Objetivo

Este documento registra a estrategia de testes que foi adicionada ao projeto apos a etapa de modernizacao.

Ele existe para responder quatro perguntas:

1. o que esta sendo testado?
2. por que separar teste unitario de teste de integracao?
3. como os stubs foram montados?
4. como rodar e interpretar a suite?

## Estado Atual

Hoje o projeto possui dois grandes grupos de testes:

- `test/unit`
- `test/integration`

Essa organizacao foi escolhida para manter a leitura didatica e para separar claramente:

- regra de negocio
- contrato HTTP

## Estrutura Criada

```text
test/
  unit/
    service_helpers_test.go
    user_service_test.go
    post_service_test.go
    comment_service_test.go
  integration/
    http_helpers_test.go
    auth_endpoints_test.go
    post_endpoints_test.go
    comment_endpoints_test.go
```

## Testes Unitarios

## Objetivo

Os testes unitarios verificam a camada `service` de forma isolada.

Ou seja:

- sem banco real
- sem HTTP real
- sem Gin como dependencia principal do teste

O foco aqui e validar regra de negocio.

## Base Compartilhada

Arquivo:

- `test/unit/service_helpers_test.go`

Esse arquivo concentra:

- stubs de `UserRepository`
- stubs de `PostRepository`
- stubs de `CommentRepository`
- helper para assert de `statusCode` a partir de `apperror`
- helper para validar claims do JWT gerado

### Por que isso e importante

Sem uma base compartilhada, cada teste repetiria:

- mocks simples
- capturas de parametros
- asserts de erro

Com isso:

- os testes ficaram menores
- a leitura ficou mais focada no comportamento
- a manutencao ficou mais simples

## `user_service_test.go`

Fluxos cobertos:

- cadastro com hash de senha
- erro de usuario ja existente
- login com refresh token existente
- login com senha invalida
- refresh token com sucesso
- refresh token com token divergente
- refresh token com falha ao apagar token antigo

### O que esse conjunto ensina

- como testar regra de negocio em Go
- como verificar side effects em stubs
- como garantir que password foi hashada
- como validar JWT sem depender do handler

## `post_service_test.go`

Fluxos cobertos:

- criacao de post
- listagem com paginacao e agregacao de comentarios
- detalhe de post inexistente
- update bloqueado por ownership
- delete com sucesso
- like/unlike com remocao de curtida existente
- erro ao consultar estado de like

### O que esse conjunto ensina

- agregacao de dados no `service`
- uso de map e slice em regra de negocio
- verificacao de autorizacao por dono do recurso
- validacao de fluxo alternativo em operacoes de toggle

## `comment_service_test.go`

Fluxos cobertos:

- criacao de comentario quando o post existe
- erro ao comentar post inexistente
- unlike de comentario ja curtido
- erro quando comentario nao existe
- erro ao remover curtida de comentario

### O que esse conjunto ensina

- coordenacao entre `commentRepo` e `postRepo`
- tratamento de cenarios principais e alternativos
- uso de stubs para garantir isolamento

## Testes de Integracao HTTP

## Objetivo

Os testes de integracao criados aqui verificam o fluxo HTTP real da aplicacao:

- rota
- bind
- validacao
- middleware
- serializacao de resposta

Eles nao integram com banco de dados.
Em vez disso, usam stubs de `service`.

### Por que isso continua sendo integracao

Porque o alvo do teste e a integracao entre componentes da camada HTTP:

- router do Gin
- middleware
- handlers
- DTOs
- helpers de resposta

## Base Compartilhada

Arquivo:

- `test/integration/http_helpers_test.go`

Esse arquivo cria:

- stubs de `UserService`
- stubs de `PostService`
- stubs de `CommentService`
- `newTestRouter(...)`
- helper para request JSON
- helper para decode de resposta
- helper para gerar token JWT valido

### O que o router de teste faz

Ele monta o `Gin` em `TestMode` e registra os handlers reais do projeto.

Assim, quando o teste faz um request com `httptest`, ele passa pelo mesmo fluxo HTTP real da aplicacao.

## `auth_endpoints_test.go`

Fluxos cobertos:

- `POST /auth/register` com sucesso
- `POST /auth/register` com erro de validacao
- `POST /auth/login` com credencial invalida
- `POST /auth/refresh` com token valido
- `POST /auth/refresh` sem header `Authorization`

### O que observar

Esses testes confirmam:

- `201` no cadastro
- `400` em validacao
- `401` em credencial invalida
- `401` em falta de autenticacao

## `post_endpoints_test.go`

Fluxos cobertos:

- `POST /tweets/` com sucesso autenticado
- `POST /tweets/` sem token
- `GET /tweets/` com `page` invalida
- `GET /tweets/:post_id/detail` com sucesso
- `GET /tweets/:post_id/detail` com `post_id` invalido

### O que observar

Esses testes validam pontos importantes da modernizacao:

- parse centralizado de params
- validacao de query params
- uso correto de `401` e `400`

## `comment_endpoints_test.go`

Fluxos cobertos:

- `POST /comment/` com sucesso autenticado
- `POST /comment/` sem token
- `POST /comment/action` retornando `404` quando comentario nao existe
- `POST /comment/action` com sucesso

### O que observar

Esses testes cobrem:

- middleware de autenticacao
- bind do corpo JSON
- propagacao de `apperror` para resposta HTTP

## O Que Foi Validado

A suite foi validada com:

```bash
go test ./...
```

E tambem com:

```bash
golangci-lint run
```

Durante a criacao da suite tambem foram verificados os diagnosticos do editor para garantir:

- imports corretos
- interfaces satisfeitas
- ausencia de erros de compilacao nos arquivos novos

## O Que Essa Suite Nao Faz

Ainda nao existem nesta rodada:

- testes de repository com banco real
- testes end-to-end subindo MySQL
- medicao de coverage no pipeline

Isso nao invalida o trabalho atual.
Na verdade, para o momento do projeto, a escolha foi intencional:

- primeiro estabilizar regra de negocio
- depois estabilizar contrato HTTP
- deixar integracao com banco como evolucao futura

## Relacao Entre Camadas e Tipos de Teste

## Unitario

Valida:

- regra
- branching
- side effects locais

Sem depender de:

- banco
- HTTP
- middleware real

## Integracao HTTP

Valida:

- roteamento
- middleware
- bind
- validacao
- status code
- payload de resposta

Sem depender de:

- banco real

## Proximo Passo Natural

Se voce quiser continuar evoluindo a maturidade do projeto, o proximo degrau em testes seria:

1. testes de repository com banco de testes
2. testes de startup e health check
3. coverage report
4. pipeline automatizado de CI

## Como Recomendo Estudar Estes Testes

Siga esta ordem:

1. `test/unit/service_helpers_test.go`
2. `test/unit/user_service_test.go`
3. `test/unit/post_service_test.go`
4. `test/unit/comment_service_test.go`
5. `test/integration/http_helpers_test.go`
6. `test/integration/auth_endpoints_test.go`
7. `test/integration/post_endpoints_test.go`
8. `test/integration/comment_endpoints_test.go`

Primeiro entenda como os stubs foram montados.
Depois veja como cada teste usa o padrao:

- arrange
- act
- assert

Se voce quiser o nivel maximo de aprofundamento, leia na sequencia:

- `17-detalhamento-completo-da-suite-de-testes.md`

## Conclusao

Com essa suite, o projeto deixou de depender apenas de validacao manual.

Agora ele passa a ensinar tambem:

- como desenhar testes em Go
- como isolar dependencias sem framework pesado
- como validar HTTP com `httptest`
- como testar services usando interfaces reais e stubs pequenos

Isso torna o `go-tweets` uma base bem mais madura para estudo de backend em Go com foco de mercado.
