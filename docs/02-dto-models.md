# DTOs e Models

## Diferenca Entre DTO e Model

Se voce vem de Java:

- `DTO` = objeto de entrada/saida da API
- `Model` = estrutura usada na camada de dominio/persistencia

Neste projeto:

- `dto` representa payloads HTTP
- `model` representa dados de banco ou composicoes usadas em repositorios

## `internal/dto/user_dto.go`

### `RegisterRequest`

```go
type RegisterRequest struct {
    Email           string `json:"email" validate:"required,email"`
    Username        string `json:"username" validate:"required,min=3"`
    Password        string `json:"password" validate:"required"`
    PasswordConfirm string `json:"password_confirm" validate:"required,eqfield=Password"`
}
```

Linha por linha:

- cada campo e uma propriedade do JSON recebido
- `json:"..."` define o nome no corpo HTTP
- `validate:"..."` define regra de validacao
- `eqfield=Password` exige que `PasswordConfirm` seja igual a `Password`

### `RegisterResponse`

- carrega apenas o `ID` criado
- isso evita devolver dados sensiveis como senha

### `LoginRequest`

- pede `email` e `password`
- a validacao exige ambos

### `LoginResponse`

- retorna `Token` e `RefreshToken`
- equivalente a um response DTO de autenticacao em Java

### `RefreshTokenRequest` e `RefreshTokenResponse`

- entrada recebe refresh token atual
- saida devolve novo access token e novo refresh token

## `internal/dto/post_dto.go`

### `CreateOrUpdatePostRequest`

- usado tanto na criacao quanto na edicao
- `Title` e `Content` sao obrigatorios

### `CreateOrUpdatePostResponse`

- devolve so o `ID` do post

### `LikeOrUnlikePostRequest`

- encapsula `post_id`
- isso simplifica bind do JSON no handler

### `Comment`

Este DTO representa comentario ja pronto para resposta HTTP.

Campos:

- `ID`
- `Username`
- `Content`
- `LikeCount`
- `CreatedAt`
- `UpdatedAt`

Observe que datas sao `string`, nao `time.Time`.
Isso ocorre porque o service converte datas antes de responder.

### `DetailPostResponse`

Representa a saida completa de um post com comentarios.

Tem:

- dados do post
- lista `Comments []Comment`

### `GetAllPostRequest`

- `Limit` e `Page`
- tags `param:"..."` indicam intencao de bind por query string

### `GetAllPostResponse`

- `TotalPage`
- `CurrentPage`
- `Limit`
- `Data []DetailPostResponse`

Isso funciona como uma pagina de resultado.

## `internal/dto/comment_dto.go`

### `StoreCommentRequest`

- recebe `post_id` e `content`
- `content` tem uma tag com espaco em `validate: "required"`, o que sugere cuidado porque tags em Go sao sensiveis a formato

### `LikeOrUnLikeCommentRequest`

- encapsula `comment_id`

## `internal/model/user_model.go`

### `UserModel`

```go
type UserModel struct {
    ID        int
    Email     string
    Username  string
    Password  string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

Leitura Java -> Go:

- seria proximo de uma entidade simples
- nao ha getters/setters
- acesso direto aos campos

### `RefreshTokenModel`

- guarda token de renovacao
- `ExpiredAt` controla validade
- `CreatedAt` e `UpdatedAt` controlam auditoria

## `internal/model/post_model.go`

### `PostModel`

- representa um post cru
- tem `ID`, `UserID`, `Title`, `Content`
- usa `time.Time` para datas

### `PostLikeModel`

- representa linha da tabela `post_likes`

### `PostWithUserModel`

Este model e interessante.
Ele nao corresponde exatamente a uma tabela unica.
Ele representa o resultado de um `JOIN`.

Campos extras:

- `Username`
- `LikeCount`

Isto e comum em Go quando o repository executa SQL manual e precisa ler colunas agregadas.

## `internal/model/comment_model.go`

### `CommentModel`

- representa comentario
- inclui `Username` e `LikeCount`
- novamente, esse model tambem acaba servindo para resultados enriquecidos com `JOIN`

### `CommentLikeModel`

- representa curtida em comentario

## Diferencas Importantes Para Quem Vem de Java

### 1. Sem anotacoes de ORM

Aqui nao existe JPA/Hibernate.
Esses `model`s nao tem `@Entity`, `@Column`, `@Id`.
O mapeamento e feito manualmente no SQL do repository.

### 2. Sem encapsulamento classico

Em Java voce veria:

```java
private String email;
public String getEmail() { ... }
```

Em Go, o estilo idiomatico geralmente usa campos exportados diretamente em structs simples.

### 3. Tipagem de tempo

No banco e repository usa-se `time.Time`.
Na resposta HTTP o codigo frequentemente converte para `string`.

### 4. Reaproveitamento de structs

No projeto, alguns `model`s servem tanto para insert quanto para retorno de join.
Em projetos maiores isso costuma ser separado de forma mais estrita.

## Resumo Pratico

- `dto` = conversa com HTTP
- `model` = conversa com service/repository/banco
- tags = controlam JSON e validacao
- `time.Time` = tipo padrao de data/hora em Go
