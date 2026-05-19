# Middleware, JWT e Utilitarios

## `internal/middleware/middleware.go`

Este arquivo contem os middlewares de autenticacao.
Se voce vem de Java, pense nele como um filtro HTTP.

## `AuthMiddleware`

### Linha 14

```go
func AuthMiddleware(secretkey string) gin.HandlerFunc {
```

- recebe a chave secreta JWT
- devolve uma funcao middleware do Gin

Em Java isso lembraria um filtro configurado com dependencia externa.

### Linha 15

```go
return func(c *gin.Context){
```

- retorna uma closure
- `c *gin.Context` representa a request atual

### Linhas 16-20

```go
header := c.Request.Header.Get("Authorization")
if header == ""{
    c.AbortWithError(http.StatusUnauthorized, errors.New("missing token"))
    return
}
```

- le header `Authorization`
- se estiver vazio, aborta com `401`
- `AbortWithError` interrompe o pipeline

### Linhas 22-26

```go
userID, username, err := jwt.ValidadeToken(header, secretkey, true)
if err != nil{
    c.AbortWithError(http.StatusUnauthorized, errors.New("missing token"))
    return
}
```

- chama funcao que valida token
- retorna `userID`, `username` e erro
- `true` indica que claims devem ser validadas

### Linhas 27-29

```go
c.Set("userID", userID)
c.Set("username", username)
c.Next()
```

- injeta dados no contexto da request
- `c.Next()` deixa a requisicao seguir

Isso lembra guardar dados num request attribute ou security context.

## `AuthRefreshTokenMiddleware`

Estrutura muito parecida com a anterior.
A diferenca principal esta nesta chamada:

```go
jwt.ValidadeToken(header, secretkey, false)
```

- `false` significa que o parse ignora validacao de claims
- isso faz sentido para refresh token em alguns fluxos, mas exige cuidado de seguranca

## `pkg/internalsql/jwt/jwt.go`

Este arquivo cria e valida JWT.

## `CreateToken`

### Linha 10

```go
func CreateToken(id int, username, secretKey string) (string, error) {
```

- recebe id, username e chave secreta
- devolve token assinado e erro

### Linhas 11-16

```go
token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
    "id":       id,
    "username": username,
    "exp":      time.Now().Add(60 * time.Minute).Unix(),
})
```

- cria token com algoritmo HS256
- `jwt.MapClaims` e um mapa de claims
- `exp` define expiracao

Comparando com Java:

- lembra criar um JWT com claims customizadas
- aqui tudo e mais manual e explicito

### Linhas 18-21

```go
key := []byte(secretKey)
tokenStr, err := token.SignedString(key)
return tokenStr, err
```

- converte secret para bytes
- assina o token
- retorna string final

## `ValidadeToken`

Apesar do nome ter typo, a intencao e "ValidateToken".

### Linhas 23-29

- prepara variaveis locais
- `claims = jwt.MapClaims{}`
- `token *jwt.Token`
- `err error`

### Linhas 31-39

- escolhe entre validacao normal e parse sem validacao de claims
- quando `withClaimValidate` e `true`, usa parse padrao
- quando `false`, usa `jwt.WithoutClaimsValidation()`

### Linhas 40-46

- se ocorrer erro, retorna
- se `token.Valid` for falso, retorna erro

### Linhas 48-60

```go
idFloat, ok := claims["id"].(float64)
...
username, ok := claims["username"].(string)
...
return int(idFloat), username, nil
```

- ao ler claims, numeros costumam vir como `float64`
- por isso ha cast para `float64` antes de converter para `int`
- depois extrai `username`

Esse trecho e muito educativo para quem vem de Java porque mostra type assertion em Go:

```go
valor, ok := algo.(Tipo)
```

Se `ok` for falso, o cast falhou.

## `pkg/internalsql/refreshtoken/refresh_token.go`

Este utilitario cria refresh tokens aleatorios.

### Linhas 8-15

```go
func GenerateRefreshToken() (string, error) {
    b := make([]byte, 18)
    _, err := rand.Read(b)
    if err != nil {
        return "", err
    }
    return hex.EncodeToString(b), nil
}
```

Explicando:

- `make([]byte, 18)` cria slice de 18 bytes
- `rand.Read(b)` preenche com bytes criptograficamente aleatorios
- `hex.EncodeToString(b)` transforma em string legivel

Em Java isso seria algo na linha de `SecureRandom` + encoding.

## Padroes de Go Que Aparecem Aqui

### 1. Closure

Middleware retorna funcao dentro de funcao.

### 2. `[]byte`

Muito usado quando lidamos com criptografia, assinatura e IO.

### 3. Type assertion

```go
idFloat, ok := claims["id"].(float64)
```

- parecido com cast dinamico, mas seguro

### 4. `make`

`make` cria slices, maps e channels.

Exemplo:

```go
b := make([]byte, 18)
```

## Fluxo Completo de Auth

1. login chama `CreateToken`
2. token vai para o cliente
3. cliente envia em `Authorization`
4. middleware chama `ValidadeToken`
5. middleware injeta `userID` e `username` no contexto
6. handler autenticado usa `c.GetInt("userID")`
