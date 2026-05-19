# Fluxo de Comentarios

## Visao Geral

O dominio `comment` cuida de:

- criar comentario em tweet
- curtir/descurtir comentario
- buscar comentarios por post

Fluxo:

```text
/comment -> handler/comment -> service/comment -> repository/comment
```

## `internal/handler/comment/handler.go`

### Estrutura

- `api *gin.Engine`
- `validate *validator.Validate`
- `commentService comment.CommentService`

### `RouteList`

```go
routeAuth := h.api.Group("/comment")
routeAuth.Use(middleware.AuthMiddleware(secretKey))
routeAuth.POST("/",h.CreateComment)
routeAuth.POST("/action", h.LikeOrUnlikecomment)
```

- todas as rotas de comentario exigem autenticacao
- `POST /comment/` cria comentario
- `POST /comment/action` faz toggle de like

## `store_comment.go` do handler

### Fluxo

1. cria `ctx`
2. faz bind de `StoreCommentRequest`
3. valida
4. pega `userID` do contexto
5. chama `CreateComment`
6. responde status

### Ponto didatico

```go
req dto.StoreCommentRequest
```

- declaracao sem `new`
- em Go isso cria a variavel zero-value da struct

## `like_or_unlike_comment.go` do handler

Segue o mesmo padrao do handler de like de post:

- bind de `LikeOrUnLikeCommentRequest`
- validacao
- leitura de `userID`
- chamada do service

## `internal/service/comment/service.go`

### Interface

```go
type CommentService interface {
    CreateComment(...)
    LikeOrUnLikeComment(...)
}
```

- define operacoes publicas da camada

### Struct concreta

```go
type commentService struct {
    cfg *config.Config
    validate *validator.Validate
    commentRepo comment.CommentRepository
    postRepo post.PostRepository
}
```

Observe:

- o service de comentario depende do `postRepo`
- isso existe porque, antes de comentar, ele valida se o tweet existe

## `create_comment.go` do service

### Linhas 13-21

```go
postExist, err := s.postRepo.GetPostByID(ctx, req.PostID)
...
if postExist == nil{
    return http.StatusNotFound, errors.New("tweet not found")
}
```

- antes de inserir comentario, valida existencia do post

### Linhas 23-30

- pega `time.Now()`
- monta `model.CommentModel`
- preenche `PostID`, `UserID`, `Content`, timestamps

### Linhas 31-35

- chama `commentRepo.StoreComment`
- retorna `201`

## `Like_or_unlike_comment.go` do service

### Linhas 13-19

- verifica se comentario existe com `DetailComment`

### Linhas 20-23

- pergunta se usuario ja curtiu

### Linhas 25-42

Logica de toggle:

- se ja curtiu: `DeleteLikeComment`
- se nao curtiu: cria `CommentLikeModel` e chama `StoreLikeComment`

### Linha 44

- retorna `200 OK`

## `internal/repository/comment/repository.go`

### Interface

Metodos expostos:

- salvar comentario
- detalhar comentario
- verificar curtida
- deletar curtida
- salvar curtida
- listar comentarios por varios posts

### Struct concreta

- armazena `db *sql.DB`

## `store_comment.go`

Query:

```sql
INSERT INTO comments (post_id, user_id, content, created_at, updated_at)
VALUES (?, ?, ?, ?, ?)
```

- insert simples
- `ExecContext` executa sem retorno de linhas

## `detail_comment.go`

Query:

```sql
SELECT id, post_id, user_id, content, created_at, updated_at
FROM comments
WHERE id = ?
```

### Padrao de retorno

- se `sql.ErrNoRows`, retorna `nil, nil`
- senao escaneia para `CommentModel`

## `is_user_already_like_comment.go`

### Query

```sql
SELECT id FROM comment_likes
WHERE comment_id = ?
AND user_id = ?
```

- se encontrar linha, retorna `true`
- se nao encontrar, retorna `false`

## `store_like_comment.go`

Query:

```sql
INSERT INTO comment_likes ( comment_id, user_id, created_at, updated_at)
VALUES(?,?,?,?)
```

- grava nova curtida

## `delete_like_comment.go`

Intencao da query:

```sql
DELETE FROM comment_likes
WHERE comment_id = ?
AND user_id = ?
```

No codigo atual ha um typo em `WEHERE`.
Isso e util para voce aprender uma licao importante em Go sem ORM:

- como o SQL fica em string, erros de digitacao so aparecem em runtime

## `get_all_comments.go`

Este e o repository mais rico do modulo.

### Objetivo

Receber varios `postIDs` e devolver todos os comentarios desses posts.

### Linhas 12-15

```go
if len(postIDs) == 0 {
    return []model.CommentModel{}, nil
}
```

- evita montar SQL invalido com `IN ()`

### Linhas 16-21

- cria lista de placeholders `?`
- cria slice `args` com os valores

Exemplo:

- entrada: `[1, 2, 3]`
- placeholders: `"?,?,?"`

### Linha 22

`fmt.Sprintf` monta a query dinamicamente com quantidade correta de placeholders.

### Query

```sql
SELECT c.id, c.post_id, c.user_id, u.username, c.content, c.created_at, c.updated_at, COUNT(cl.id) as like_count
FROM comments as c
JOIN users as u ON u.id = c.user_id
LEFT JOIN comment_likes as cl ON cl.comment_id = c.id
WHERE c.post_id IN (%s)
GROUP BY c.id, c.post_id, c.user_id, u.username, c.content, c.created_at, c.updated_at
ORDER BY like_count DESC
```

Interpretacao:

- pega comentarios
- traz username do autor
- conta curtidas por comentario
- filtra por varios posts ao mesmo tempo

### Linhas 30-35

- executa `QueryContext`
- se nao houver linhas, devolve slice vazio

### Linhas 37-56

- percorre resultado com `rows.Next()`
- faz `rows.Scan(...)`
- monta `[]model.CommentModel`

## Como os Comentarios Encaixam no Fluxo de Posts

O modulo de post usa `commentRepo.GetCommentsByPostIDs(...)` em dois cenarios:

- `DetailPost`: busca comentarios de um unico post
- `GetAllPost`: busca comentarios dos posts da pagina atual

Isto e uma forma eficiente de evitar uma query por post na listagem.

## Conceitos de Go Que Ficam Claros Neste Modulo

### 1. Slice

```go
[]model.CommentModel
```

- equivalente a uma lista tipada

### 2. Map

Nos services de post, comentarios sao agrupados em:

```go
map[int][]dto.Comment
```

- equivalente a `Map<Integer, List<Comment>>`

### 3. Zero values

Structs e slices sao usados sem construtores complexos.

### 4. SQL manual

Sem Hibernate, Criteria API ou JPQL.
Tudo fica explicito no repository.
