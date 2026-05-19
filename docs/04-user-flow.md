# Fluxo de Usuario: Handler, Service e Repository

## Visao Geral

O dominio `user` cuida de:

- registro
- login
- refresh token

Fluxo:

```text
HTTP /auth/* -> handler/user -> service/user -> repository/user -> MySQL
```

## `internal/handler/user/handler.go`

### Estrutura `Handler`

```go
type Handler struct{
    api *gin.Engine
    validate *validator.Validate
    userService user.UserService
}
```

- `api`: roteador principal
- `validate`: validador de DTO
- `userService`: dependencia da regra de negocio

### `NewHandler`

- padrao de construtor em Go
- retorna `*Handler`

### `RouteList`

```go
authRoute := h.api.Group("/auth")
authRoute.POST("/register",h.Register)
authRoute.POST("/login",h.Login)
```

- cria grupo `/auth`
- registra rotas publicas

```go
refreshRoute := h.api.Group("/auth")
refreshRoute.Use(middleware.AuthRefreshTokenMiddleware(secretKey))
refreshRoute.POST("/refresh",h.RefreshToken)
```

- cria outro grupo `/auth`
- adiciona middleware de refresh
- protege a rota `/refresh`

## `internal/handler/user/register.go`

### Linhas 9-13

- cria `ctx` da request
- declara `req dto.RegisterRequest`

### Linhas 14-19

- `ShouldBindJSON(&req)` tenta ler JSON do body
- se falhar, retorna `400`

### Linhas 21-25

- executa validacao das tags do DTO
- se falhar, responde com `400`

### Linhas 29-40

- chama `h.userService.Register`
- recebe `userID`, `statusCode` e `err`
- em caso de sucesso devolve `dto.RegisterResponse`

Note um padrao muito comum no projeto:

```go
resultado, statusCode, err := service.Algo(...)
```

Isso substitui o uso de excecoes com mapeamento global de erro.

## `internal/handler/user/login.go`

Fluxo praticamente identico ao register:

1. le JSON
2. valida DTO
3. chama service
4. transforma em JSON de resposta

Diferenca:

- a resposta e `dto.LoginResponse`
- devolve `Token` e `RefreshToken`

## `internal/handler/user/refresh_token.go`

### Ponto importante

```go
userID := c.GetInt("userID")
```

- esse valor nao vem do body
- ele foi colocado pelo middleware JWT

Ou seja:

- middleware valida token
- handler reaproveita dados do contexto

## `internal/service/user/service.go`

Aqui ficam interface e implementacao.

### `UserService interface`

```go
type UserService interface{
    Register(...)
    Login(...)
    RefreshToken(...)
}
```

- define contrato da camada
- ajuda a desacoplar handler de implementacao concreta

### `userService struct`

- guarda `cfg`
- guarda `userRepo`

### `NewUserService`

- injeta dependencias
- retorna `*userService`

## `internal/service/user/register.go`

Este e o coracao da regra de cadastro.

### Linhas 16-22

```go
userExist, err := s.userRepo.GetUserByEmailOrUsername(ctx, req.Email, req.Username)
...
if userExist != nil{
    return 0, http.StatusBadRequest, errors.New("user already exists")
}
```

- verifica duplicidade
- se ja existe, interrompe

### Linhas 24-27

```go
passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
```

- faz hash da senha
- em Java isso lembraria chamar BCryptPasswordEncoder

### Linhas 29-36

- cria `model.UserModel`
- popula email, username, password hash e timestamps

### Linhas 37-41

- chama `CreateUser`
- devolve `userID`

## `internal/service/user/login.go`

### Linhas 18-25

- busca usuario por email
- se nao encontrar, devolve erro generico de login

### Linhas 27-30

```go
err = bcrypt.CompareHashAndPassword([]byte(userExist.Password),[]byte(req.Password))
```

- compara hash salvo com senha enviada

### Linhas 32-35

- gera JWT com `jwt.CreateToken`

### Linhas 37-45

- verifica se ja existe refresh token valido
- se existir, reaproveita

### Linhas 47-58

- gera novo refresh token aleatorio
- monta `RefreshTokenModel`
- persiste no banco

### Linha 64

- retorna `token`, `refreshToken` e `200`

## `internal/service/user/refresh_token.go`

Fluxo:

1. busca usuario por ID
2. verifica se existe refresh token ativo
3. compara token enviado com token salvo
4. gera novo access token
5. remove refresh token antigo
6. cria novo refresh token

### Linhas 17-23

- garante que o usuario existe

### Linhas 25-34

- valida refresh token
- precisa existir, nao estar expirado e ser igual ao enviado

### Linhas 36-39

- cria novo JWT

### Linhas 41-44

- remove refresh token antigo

### Linhas 46-58

- gera e salva refresh token novo

Observacao didatica:

- o metodo chama `s.userRepo.StoreRefreshToken(...)` sem checar erro no retorno final
- em termos de logica, a intencao e boa, mas essa ausencia de checagem merece atencao em projetos reais

## `internal/repository/user/repository.go`

Este arquivo define a interface do repository e a struct concreta.

### Interface

Metodos principais:

- buscar usuario por email/username
- criar usuario
- buscar refresh token
- salvar refresh token
- buscar usuario por ID
- deletar refresh token

### Implementacao concreta

```go
type userRepository struct {
    db *sql.DB
}
```

- guarda o pool de conexao

## `create_user.go`

### Query

```sql
INSERT INTO users (email, username, password, created_at, updated_at)
VALUES (?,?,?,?,?)
```

### Execucao

- `ExecContext` executa insert
- `LastInsertId()` recupera o ID gerado

## `get_user_by_email_username.go`

### Query

```sql
SELECT id, username, email, password, created_at, updated_at
FROM users
WHERE email = ?
or username = ?
```

### Detalhes

- `QueryRowContext` busca uma linha
- `row.Scan(...)` mapeia colunas para campos
- se `sql.ErrNoRows`, retorna `nil, nil`

Esse padrao aparece muito em Go:

- `nil, nil` = nao encontrou, mas nao houve erro tecnico

## `get_user_by_id.go`

- busca usuario por ID
- parecido com o metodo anterior
- nao busca password

## `store_refresh_token.go`

- insere refresh token na tabela `refresh_tokens`
- usa `ExecContext`

## `get_refresh_token.go`

### Query

```sql
SELECT id, user_id, refresh_token, expired_at
FROM refresh_tokens
WHERE user_id = ? AND expired_at > ?
```

- devolve token ainda valido
- `now time.Time` entra como parametro da query

## `delete_refresh_token.go`

- executa `DELETE`
- depois usa `RowsAffected()` para verificar se algo foi apagado
- se `0`, retorna erro `"nothing to delete"`

## Resumo do Fluxo de Usuario

### Cadastro

1. handler valida request
2. service verifica duplicidade
3. service faz hash da senha
4. repository executa insert

### Login

1. handler valida request
2. service busca usuario
3. service compara senha
4. service gera access token
5. service busca ou cria refresh token

### Refresh

1. middleware extrai usuario do token
2. handler pega `userID`
3. service valida refresh token salvo
4. service troca token antigo por novo
