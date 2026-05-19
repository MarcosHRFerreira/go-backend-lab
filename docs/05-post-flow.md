# Fluxo de Posts: Handler, Service e Repository

## Visao Geral

O dominio `post` implementa:

- criar tweet
- atualizar tweet
- excluir tweet logicamente
- curtir/descurtir tweet
- detalhar tweet
- listar tweets com paginacao

Fluxo padrao:

```text
/tweets -> handler/post -> service/post -> repository/post
```

## `internal/handler/post/handler.go`

### Estrutura

- guarda `api`, `validate` e `postService`

### `RouteList`

Rotas autenticadas:

- `POST /tweets/` -> `CreatePost`
- `PUT /tweets/:post_id/update` -> `UpdatePost`
- `DELETE /tweets/:post_id/delete` -> `DeletePost`
- `POST /tweets/action` -> `LikeOrUnlikePost`

Rotas publicas:

- `GET /tweets/:post_id/detail` -> `DetailPost`
- `GET /tweets/` -> `GetallPost`

Observe a combinacao de grupos com e sem middleware.

## Handlers de Post

Os handlers seguem sempre o mesmo template:

1. obter `ctx`
2. bind do JSON ou param/query
3. validar entrada
4. ler `userID` do contexto quando necessario
5. chamar `postService`
6. responder JSON

## `create_post.go`

### Linhas 11-14

- cria contexto da requisicao
- instancia `dto.CreateOrUpdatePostRequest`

### Linhas 15-20

- faz bind do body JSON
- se falhar, responde `400`

### Linhas 22-27

- valida o DTO

### Linha 30

```go
userID := c.GetInt("userID")
```

- pega usuario autenticado do middleware

### Linha 31

```go
postID, statusCode, err := h.postService.CreatePost(ctx, &req, userID)
```

- chama regra de negocio

### Linhas 39-41

- responde com `CreateOrUpdatePostResponse`

## `update_post.go`

### Diferencas principais

- le `post_id` da URL com `c.Param("post_id")`
- converte string para inteiro com `strconv.ParseInt`
- passa `postIDInt` ao service

Em Go e comum converter explicitamente tipos numericos.
Java faz muito disso implicitamente menos vezes; Go exige clareza.

## `delete_post.go`

- le `userID` do contexto
- le `post_id` da rota
- converte com `strconv.Atoi`
- chama `DeletePost`
- devolve mensagem de sucesso

## `detail_post.go`

- rota publica
- le `post_id`
- chama `DetailPost`
- devolve um `DetailPostResponse`

## `get_all_post.go` do handler

### Linhas importantes

```go
pageStr := c.DefaultQuery("page", "1")
limitStr := c.DefaultQuery("limit", "2")
```

- le query params
- usa valores padrao caso a query nao venha

```go
page, _ := strconv.ParseInt(pageStr, 10, 64)
limit, _ := strconv.ParseInt(limitStr, 10, 64)
```

- converte para inteiro
- o `_` ignora erro

Se voce vem de Java, o `_` em Go significa "nao me importo com este valor".

Depois:

- monta `dto.GetAllPostRequest`
- chama o service
- responde JSON

## `like_or_unlike_post.go` do handler

- faz bind de `LikeOrUnlikePostRequest`
- valida
- usa `userID` do middleware
- chama `LikeOrUnlikePost`

## `internal/service/post/service.go`

### Interface

Define operacoes do dominio `post`.

### `postService`

Guarda:

- `cfg`
- `postRepo`
- `commentRepo`

O `commentRepo` aparece porque detalhe e listagem de posts tambem carregam comentarios.

## `create_post.go` do service

### Fluxo

1. pega `time.Now()`
2. monta `model.PostModel`
3. chama `StorePost`
4. devolve `201 Created`

### Ponto didatico

```go
now := time.Now()
```

- `time.Now()` e o equivalente ao `LocalDateTime.now()`/`Instant.now()`

## `update_post.go` do service

### Linhas 13-19

- busca post por ID
- se nao existir, retorna `404`

### Linhas 21-24

- verifica se o `userID` do dono e o mesmo do usuario autenticado
- se nao for, retorna "tweet not found"

### Linhas 25-32

- monta model parcial com titulo, conteudo e `UpdatedAt`
- chama `UpdatePost` do repository

## `delete_post.go` do service

Fluxo quase igual ao `UpdatePost`:

1. verifica existencia do tweet
2. verifica ownership
3. chama `SoftDeletePost`

## `like_or_unlike_post.go` do service

### Ideia central

O metodo faz toggle:

- se usuario ja curtiu, remove a curtida
- se nao curtiu, cria curtida

### Linhas 15-21

- garante que o post existe

### Linhas 23-26

- pergunta ao repository se o usuario ja curtiu

### Linhas 28-45

- se sim: `DeleteLikePost`
- se nao: cria `PostLikeModel` e chama `StoreLikePost`

## `detail_post.go` do service

Este metodo combina dados de post com comentarios.

### Linhas 14-20

- busca post por ID
- se nao existe, devolve `404`

### Linha 22

```go
comments, err := s.commentRepo.GetCommentsByPostIDs(ctx, []int{postID})
```

- usa slice literal `[]int{postID}`
- isso cria uma lista com um unico elemento

### Linhas 26-35

- converte `[]model.CommentModel` para `[]dto.Comment`
- transforma `time.Time` em string

### Linhas 37-45

- monta resposta detalhada do post

## `get_all_post.go` do service

Este e um dos metodos mais ricos do projeto.

### Linhas 14-17

- chama `TotalPost`
- usa isso para calcular total de paginas

### Linha 19

```go
offset := param.Limit * (param.Page -1)
```

- calcula deslocamento da paginacao
- mesma ideia de `offset` em SQL/JPA

### Linhas 20-23

- carrega posts da pagina atual

### Linhas 25-28

```go
postIDs := make([]int, len(posts))
for _, post := range posts{
    postIDs = append(postIDs, post.ID)
}
```

Didaticamente:

- `make([]int, len(posts))` cria slice ja com tamanho `len(posts)`
- depois `append` adiciona novos itens

Ou seja, a intencao e montar lista de IDs, mas esse padrao deixa elementos zero no inicio.
Para aprender Go, o mais comum aqui seria:

```go
postIDs := make([]int, 0, len(posts))
```

### Linhas 29-32

- busca todos os comentarios dos posts da pagina

### Linhas 34-44

- monta `commentsMap := make(map[int][]dto.Comment)`
- agrupa comentarios por `PostID`

Esse `map[int][]dto.Comment` lembra um `Map<Integer, List<Comment>>` do Java.

### Linhas 46-58

- percorre posts
- monta `[]dto.DetailPostResponse`
- para cada post injeta `commentsMap[post.ID]`

### Linha 60

```go
totalPage := int64(math.Ceil(float64(totalPost) / float64(param.Limit)))
```

- converte inteiros para `float64`
- faz divisao com decimal
- aplica `math.Ceil`
- volta para `int64`

Isso e necessario porque Go nao mistura tipos numericos automaticamente.

## `internal/repository/post/repository.go`

### Interface

Expose:

- insert
- buscar por ID
- update
- soft delete
- verificar curtida
- deletar curtida
- salvar curtida
- contar posts
- listar posts

### Struct concreta

```go
type postRepository struct {
    db *sql.DB
}
```

- recebe conexao com banco

## `store_post.go`

- executa `INSERT INTO posts`
- usa `ExecContext`
- retorna `LastInsertId`

## `update_post.go`

- faz `UPDATE posts SET title = ?, content = ?, updated_at = ?`
- verifica `RowsAffected`
- se nao atualizou nada, retorna erro

## `soft_delete_post.go`

Soft delete significa:

- nao remove a linha
- apenas preenche `delete_at`

Query:

```sql
UPDATE posts SET delete_at = ? WHERE id = ?
```

## `is_user_already_post.go`

Apesar do nome do arquivo nao estar ideal, a funcao verifica se o usuario ja curtiu o post.

### Logica

- faz `SELECT id FROM post_likes WHERE post_id = ? AND user_id = ?`
- se `sql.ErrNoRows`, retorna `false, nil`
- senao, `true, nil`

## `store_like_post.go`

- insere na tabela `post_likes`

## `delete_like_post.go`

- remove da tabela `post_likes`

## `get_post_by_id.go`

Este repository ja faz `JOIN` com `users` e `post_likes`.

### Query principal

```sql
SELECT p.id, p.title, p.content, p.user_id, p.created_at, p.updated_at, u.username, COUNT(pl.id) as like_count
FROM posts as p
JOIN users as u ON p.user_id = u.id
LEFT JOIN post_likes as pl ON pl.post_id = p.id
WHERE p.id = ?
AND p.delete_at IS NULL
GROUP BY p.id, p.title, p.content, p.user_id, p.created_at, p.updated_at, u.username
```

Interpretacao:

- pega dados do post
- junta dono do post
- conta curtidas
- ignora posts deletados logicamente

Depois:

- `row.Scan(...)` joga tudo em `PostWithUserModel`

## `get_all_post.go` do repository

Mesma ideia do `get_post_by_id`, mas para pagina.

### Query

- seleciona campos do post
- junta usuario
- junta curtidas
- filtra `p.delete_at IS NULL`
- agrupa
- ordena por data desc
- usa `LIMIT` e `OFFSET`

### `rows.Next()`

```go
for rows.Next(){
```

- itera sobre resultado do banco
- equivalente a percorrer `ResultSet` em JDBC

### `rows.Scan(...)`

- mapeia colunas em variaveis do model

## `total_post.go`

Query simples:

```sql
SELECT COUNT(id) FROM posts WHERE delete_at IS NULL
```

- devolve quantidade de tweets nao deletados

## Resumo Mental Para Quem Vem de Java

### Handler

- parecido com Controller

### Service

- parecido com Service do Spring
- concentra regra

### Repository

- parecido com DAO/Repository
- mas sem ORM
- SQL escrito na mao

### Diferenca principal

Em Go, o fluxo e explicito.
Nao ha annotations escondendo comportamento.
Voce enxerga o caminho completo do request ate o SQL.
