# Go Tweets: Visao Geral Para Quem Vem de Java

## Objetivo do Projeto

Este projeto e uma API HTTP escrita em Go usando Gin e MySQL.
Ele implementa um pequeno sistema de tweets com:

- cadastro e login de usuario
- emissao de JWT e refresh token
- criacao, edicao e exclusao logica de posts
- curtidas em posts
- comentarios e curtidas em comentarios
- listagem paginada e detalhe de tweet

Se voce vem de Java, pense nesta divisao:

- `handler` = controller
- `service` = service
- `repository` = repository/DAO
- `model` = entidades de persistencia
- `dto` = request/response objects
- `middleware` = filtros/interceptors
- `pkg` = utilitarios compartilhados
- `cmd/main.go` = ponto de entrada da aplicacao

## Fluxo Geral da Aplicacao

O fluxo padrao de uma requisicao e:

1. O cliente envia uma requsicao HTTP.
2. O `handler` recebe a rota, faz bind do JSON e validacao.
3. O `service` executa regras de negocio.
4. O `repository` acessa o MySQL com SQL manual.
5. O resultado volta para o `service`.
6. O `handler` transforma isso em resposta HTTP JSON.

Em pseudo-fluxo:

```text
HTTP -> Gin Router -> Handler -> Service -> Repository -> MySQL
```

## Mapa de Pastas

- `cmd/`
  - inicializa servidor, dependencias e rotas
- `internal/config/`
  - carrega `.env`
- `internal/dto/`
  - structs usadas na entrada e saida da API
- `internal/model/`
  - structs do dominio/persistencia
- `internal/handler/`
  - camada HTTP
- `internal/service/`
  - regras de negocio
- `internal/repository/`
  - SQL e acesso ao banco
- `internal/middleware/`
  - autenticacao
- `pkg/internalsql/`
  - conexao MySQL, JWT e refresh token
- `db/migrations/`
  - definicao do schema

## Conceitos de Go Importantes Para Quem Vem de Java

### 1. `package`

Em Go, cada arquivo pertence a um `package`.
Isso lembra o namespace/pacote do Java, mas com regras mais simples.

Exemplo:

```go
package post
```

Isso significa que o arquivo faz parte do pacote `post`.

### 2. `import`

`import` declara dependencias do arquivo.
Em Java seria equivalente aos `import` normais, mas Go organiza isso de forma bem objetiva.

### 3. `struct`

`struct` em Go lembra uma classe simples com campos publicos.
Ela nao possui construtores obrigatorios nem heranca.

Exemplo:

```go
type UserModel struct {
    ID       int
    Email    string
    Username string
}
```

Se o nome do campo comeca com letra maiuscula, ele e exportado.
Se comeca com minuscula, ele e privado ao pacote.

### 4. Metodos com receiver

Em vez de escrever:

```java
class UserService {
    int register(...) { ... }
}
```

Em Go voce escreve:

```go
func (s *userService) Register(...) (...) {
}
```

`(s *userService)` e o receiver, equivalente ao `this` do Java.

### 5. Interfaces

Interfaces em Go sao pequenas e baseadas em comportamento.
Voce nao precisa declarar explicitamente `implements`.
Se um tipo tem os metodos certos, ele satisfaz a interface.

### 6. Ponteiros

`*Tipo` significa ponteiro para aquele tipo.
Isso e parecido com passar referencia em Java, mas em Go isso e explicito.

Exemplo:

```go
func LoadConfig() (*Config, error)
```

O retorno e um ponteiro para `Config`.

### 7. Retorno multiplo

Go retorna multiplos valores com frequencia.
Isto substitui excecoes em muitas situacoes.

Exemplo:

```go
user, err := repo.GetUserByID(ctx, id)
```

Aqui temos:

- `user`: resultado
- `err`: erro

### 8. Tratamento de erro

Go nao usa `try/catch` como estilo principal.
O padrao e:

```go
if err != nil {
    return nil, err
}
```

Isso aparece em praticamente todo o projeto.

### 9. `context.Context`

Quase todas as operacoes passam `ctx context.Context`.
Pense nisso como um objeto de contexto da requisicao que carrega:

- cancelamento
- deadline
- metadados da requisicao

Em Java, a ideia se aproxima de carregar dados da request ou de um contexto de execucao.

### 10. Tags em structs

Go usa tags para JSON e validacao:

```go
Email string `json:"email" validate:"required,email"`
```

Isso diz:

- nome do campo no JSON: `email`
- regras de validacao: obrigatorio e formato de email

## Injeccao de Dependencia Neste Projeto

Este projeto nao usa framework de DI.
Tudo e montado manualmente no `main.go`.

Fluxo:

1. cria `repository`
2. injeta repository no `service`
3. injeta service no `handler`
4. registra rotas

Se voce vem de Spring Boot, pense como um wiring manual sem container IoC.

## Convencoes Que Voce Vai Ver Muito

- `NewXxx(...)` cria uma instancia
- `ShouldBindJSON` le JSON da request
- `validate.Struct(req)` executa validacao
- `c.JSON(...)` escreve resposta HTTP
- `QueryRowContext` busca uma linha
- `ExecContext` executa insert/update/delete
- `sql.ErrNoRows` sinaliza "nao encontrado"

## Ordem Recomendada de Leitura

Para entender o projeto com mais facilidade:

1. `docs/01-bootstrap-config.md`
2. `docs/02-dto-models.md`
3. `docs/03-auth-and-utils.md`
4. `docs/04-user-flow.md`
5. `docs/05-post-flow.md`
6. `docs/06-comment-flow.md`
7. `docs/07-database-and-migrations.md`

## Observacao Importante

O projeto tem alguns pontos com inconsistencias de nomenclatura e pequenos bugs que foram aparecendo durante a evolucao.
Na documentacao eu explico tanto a intencao quanto o comportamento real do codigo atual, para te ajudar a aprender Go e tambem a enxergar problemas de implementacao.
