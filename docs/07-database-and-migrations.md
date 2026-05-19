# Banco de Dados e Migrations

## Visao Geral do Schema

O banco tem as tabelas:

- `users`
- `posts`
- `comments`
- `post_likes`
- `comment_likes`
- `refresh_tokens`

Relacoes principais:

- um `user` cria muitos `posts`
- um `user` cria muitos `comments`
- um `post` tem muitos `comments`
- `post_likes` liga usuario a post
- `comment_likes` liga usuario a comentario
- `refresh_tokens` liga usuario a token de renovacao

## Como Ler as Migrations

Cada arquivo tem duas partes:

- `-- migrate:up` = aplica a mudanca
- `-- migrate:down` = desfaz a mudanca

Pense nisso como versionamento incremental do schema.

## `20260502134636_create_user_table.sql`

### Up

```sql
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(250) NOT NULL UNIQUE,
    username VARCHAR(100) NOT NULL,
    password VARCHAR(500) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

Explicando:

- `id`: chave primaria auto incremental
- `email`: unico
- `username`: nome visivel do usuario
- `password`: hash da senha
- `created_at`: data de criacao
- `updated_at`: data de atualizacao automatica

### Down

- remove a tabela `users`

## `20260502135935_create_post_table.sql`

### Estrutura

```sql
CREATE TABLE IF NOT EXISTS posts(
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    title VARCHAR(250) NOT NULL,
    content LONGTEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    delete_at TIMESTAMP NULL,
    CONSTRAINT fk_user_id_posts FOREIGN KEY (user_id) REFERENCES users(id) 
)
```

Observacoes:

- `user_id` aponta para dono do tweet
- `delete_at` implementa exclusao logica
- o nome usado e `delete_at`, nao `deleted_at`

Esse detalhe e importante porque o codigo precisa usar o mesmo nome da coluna.

## `20260502141104_create_comment_table.sql`

### Estrutura

- `post_id` referencia `posts(id)`
- `user_id` referencia `users(id)`
- `content` guarda texto do comentario

Ou seja, comentario pertence a um post e a um usuario.

## `20260502142057_create_post_like_table.sql`

Tabela de associacao:

- `post_id`
- `user_id`
- timestamps

Ela representa "usuario X curtiu post Y".

## `20260502142854_create_comment_like_table.sql`

Mesmo raciocinio, agora para comentarios:

- `comment_id`
- `user_id`

## `20260502143528_create_refresh_token_table.sql`

### Estrutura original

```sql
CREATE TABLE IF NOT EXISTS refresh_tokens(
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    refresh_token TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_user_id_refresk_tokens FOREIGN KEY (user_id) REFERENCES users(id)
)
```

Ela guarda o refresh token por usuario.

## `20260504141918_add_expired_at_into_refresh_token_table.sql`

### Up

```sql
ALTER TABLE refresh_tokens ADD expired_at TIMESTAMP NULL AFTER refresh_token;
```

- adiciona coluna de expiracao
- importante para invalidar tokens antigos

### Down

- remove a coluna `expired_at`

## `20260502150000_rename_timestamp_columns.sql`

Esta migration e mais avancada.
Ela tenta renomear colunas antigas `create_at`/`update_at` para `created_at`/`updated_at`.

## O que ela faz conceitualmente

Para cada tabela:

1. olha `information_schema.columns`
2. verifica se a coluna antiga existe
3. se existir, executa `ALTER TABLE ... RENAME COLUMN ...`
4. se nao existir, executa `SELECT 1`

Isso torna a migration mais tolerante a estados anteriores diferentes do banco.

## Trecho tipico

```sql
SET @stmt = (
    SELECT IF(
        EXISTS (
            SELECT 1
            FROM information_schema.columns
            WHERE table_schema = DATABASE()
              AND table_name = 'users'
              AND column_name = 'create_at'
        ),
        'ALTER TABLE users RENAME COLUMN create_at TO created_at',
        'SELECT 1'
    )
);
PREPARE stmt FROM @stmt;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
```

Explicacao linha a linha:

- `SET @stmt = (...)`: monta SQL dinamico
- `EXISTS (...)`: verifica se a coluna antiga existe
- `ALTER TABLE ...`: renomeia se existir
- `SELECT 1`: no-op se nao existir
- `PREPARE/EXECUTE/DEALLOCATE`: executa SQL dinamico

No `down`, o processo e invertido.

## Relacao Entre Schema e Codigo

### `users`

Usada por:

- repository/user
- login e cadastro

### `posts`

Usada por:

- repository/post
- update/delete/detail/list

### `comments`

Usada por:

- repository/comment
- detalhes e listagem de post

### `post_likes`

Usada por:

- toggle de curtida em post
- contagem com `COUNT(pl.id)`

### `comment_likes`

Usada por:

- toggle de curtida em comentario
- contagem com `COUNT(cl.id)`

### `refresh_tokens`

Usada por:

- login
- refresh token

## Cuidados Didaticos Importantes

### 1. Nome de coluna precisa bater exatamente

Exemplo do projeto:

- migration usa `delete_at`
- se o codigo usar `deleted_at`, vai falhar

### 2. SQL manual deixa tudo visivel

Vantagem:

- voce entende exatamente o que esta sendo executado

Desvantagem:

- typos aparecem so em runtime

### 3. `COUNT` + `GROUP BY`

No projeto, likes sao agregados com:

```sql
COUNT(pl.id) as like_count
```

e agrupados por colunas do post/comentario.

Isso e equivalente a consultas agregadas que em Java poderiam aparecer em SQL nativo, JPQL ou Criteria.

## Como Recomendo Estudar Este Banco

1. leia as tabelas `users`, `posts`, `comments`
2. depois leia `post_likes` e `comment_likes`
3. por fim veja `refresh_tokens`
4. compare cada tabela com os arquivos em `internal/model`
5. depois confira como cada repository faz `SELECT`, `INSERT`, `UPDATE` e `DELETE`
