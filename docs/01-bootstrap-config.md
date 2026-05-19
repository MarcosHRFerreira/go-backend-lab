# Bootstrap, Configuracao e Entrada da Aplicacao

## `cmd/main.go`

Este arquivo e o equivalente ao `Application.java` ou `main()` principal do projeto Java.
Ele sobe o servidor HTTP, conecta no banco e registra as rotas.

### Linhas 1-2

```go
package main
```

- `main` e um package especial em Go.
- Somente um pacote `main` pode gerar executavel.

### Linhas 3-22

Bloco de imports:

- `fmt`, `log`, `net/http`: bibliotecas padrao
- `config`: leitura do `.env`
- `handler`, `service`, `repository`: camadas da aplicacao
- `internalsql`: conexao com banco
- `gin`: framework HTTP
- `validator`: validacao de DTOs

Se voce vem de Java, pense nesses imports como os `import` do Spring, mas sem annotations magicas.

### Linha 24

```go
func main() {
```

- ponto de entrada do programa
- equivalente ao `public static void main(String[] args)`

### Linha 25

```go
r := gin.Default()
```

- cria o engine HTTP do Gin
- `:=` significa "declare e inicialize"
- `r` vira o roteador principal

### Linha 26

```go
validate := validator.New()
```

- cria o validador usado pelos handlers
- funciona em conjunto com tags como `validate:"required"`

### Linhas 28-31

```go
cfg, err := config.LoadConfig()
if err != nil{
    log.Fatal(err)
}
```

- chama a funcao que le variaveis de ambiente
- `cfg` recebe a configuracao
- `err` recebe eventual erro
- `log.Fatal(err)` imprime e encerra a aplicacao

Em Java isso lembraria:

```java
Config cfg = config.load();
if (cfg == null) throw ...
```

mas em Go o padrao e retorno explicito de erro.

### Linhas 33-37

```go
db, err := internalsql.ConectMySql(cfg)
if err != nil{
    log.Fatal(err)
}
```

- abre a conexao com MySQL
- se falhar, encerra a aplicacao

### Linhas 40-41

```go
r.Use(gin.Logger())
r.Use(gin.Recovery())
```

- adiciona middlewares globais do Gin
- `Logger()`: loga requests
- `Recovery()`: evita que panic derrube o servidor

Pense como filtros globais.

### Linhas 43-47

Rota de health check:

```go
r.GET("/check-health", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "message": "it's works",
    })
})
```

- registra `GET /check-health`
- usa uma funcao anonima
- `gin.H` e um `map[string]any` conveniente para JSON
- retorna `200 OK`

### Linhas 49-51

```go
userRepo := userRepo.NewRepository(db)
postRepo := postRepo.NewPostRepository(db)
commentRepo := commentRepo.NewCommentRepository(db)
```

- cria os repositories
- cada repository recebe `db *sql.DB`
- aqui nasce a camada de acesso a dados

### Linhas 53-55

```go
userService := userService.NewUserService(cfg, userRepo)
postService := postService.NewPostService(cfg, postRepo, commentRepo)
commentService := commentService.NewCommentService(cfg, commentRepo,postRepo)
```

- cria os services
- note a injecao de dependencia manual
- `postService` depende de `postRepo` e `commentRepo`
- `commentService` depende de `commentRepo` e `postRepo`

### Linhas 57-59

```go
userHandler := userHandler.NewHandler(r,validate, userService)
postHandler := postHandler.NewHandler(r, validate, postService)
commentHandler := commentHandler.NewHandler(r, validate, commentService)
```

- cria os handlers
- handler recebe router, validador e service

### Linhas 61-63

```go
userHandler.RouteList(cfg.SecretJwt)
postHandler.RouteList(cfg.SecretJwt)
commentHandler.RouteList(cfg.SecretJwt)
```

- cada handler registra suas rotas
- a secret JWT e passada porque alguns grupos usam middleware autenticado

### Linhas 65-66

```go
server:=fmt.Sprintf("127.0.0.1:%s", cfg.Port)
r.Run(server)
```

- monta endereco do servidor
- sobe o HTTP server do Gin

## `internal/config/config.go`

Este arquivo centraliza configuracoes vindas do ambiente.
Pense nele como uma classe `AppProperties`.

### Linhas 1-9

- define o pacote
- importa `fmt`, `log`, `os` e `godotenv`
- `godotenv` carrega o arquivo `.env` para o ambiente local

### Linhas 12-22

```go
type Config struct{
    Port string
    DBUrlMigration string
    SecretJwt string
    DBHost string
    DBUser string
    DBName string
    DBPassword string
    DBPort string
}
```

- `type Config struct` define a estrutura de configuracao
- cada campo guarda um valor lido do ambiente
- todos sao `string` porque `os.Getenv` retorna string

### Linhas 24-42

```go
func LoadConfig() (*Config, error){
```

- retorna ponteiro para `Config` e erro

#### Linhas 25-28

```go
err := godotenv.Load()
if err != nil{
    return nil, fmt.Errorf("failed to load .env file")
}
```

- tenta carregar o `.env`
- se falhar, devolve erro

#### Linha 30

```go
log.Println("config loaded")
```

- apenas log informativo

#### Linhas 32-41

- cria e retorna uma instancia de `Config`
- cada campo vem de `os.Getenv(...)`

## `pkg/internalsql/mysql.go`

Este arquivo abre a conexao com o MySQL.

### Linhas 1-10

- importa `database/sql`
- importa o driver MySQL com `_ "github.com/go-sql-driver/mysql"`

Esse `_` e importante:

- ele importa so pelo efeito colateral
- registra o driver no pacote `database/sql`

Em Java seria algo proximo de registrar um driver JDBC.

### Linha 12

```go
func ConectMySql(cfg *config.Config) (*sql.DB, error){
```

- recebe a configuracao
- devolve conexao com banco ou erro

### Linha 14

```go
dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=%s", ...)
```

- monta a DSN do MySQL
- `parseTime=true` e crucial para ler `TIMESTAMP` como `time.Time`

### Linhas 17-20

```go
db, err := sql.Open("mysql", dataSourceName)
if err != nil {
    return nil, fmt.Errorf("failed to connect to database")
}
```

- cria um handle de conexao
- `sql.Open` nao necessariamente abre a conexao imediatamente; ele prepara o pool

### Linhas 22-24

- loga sucesso
- retorna `db`

## `go.mod`

Este arquivo e o equivalente ao `pom.xml` ou `build.gradle`.

### Linhas 1-3

- nome do modulo: `go-tweets`
- versao da linguagem: `go 1.26.2`

### Linhas 5-12

Dependencias diretas:

- `gin`: API HTTP
- `validator`: validacao
- `mysql driver`: driver de banco
- `jwt`: tokens
- `godotenv`: leitura de `.env`
- `x/crypto`: bcrypt

### Linhas 14-43

- dependencias transitivas/indiretas
- semelhantes as dependencias trazidas automaticamente por bibliotecas principais no ecossistema Java

## `docker-compose.yml`

Este arquivo sobe o MySQL localmente.

### Linha 1

```yaml
services:
```

- raiz dos servicos do compose

### Linhas 2-15

Servico `db`:

- imagem `mysql:8.0`
- nome do container `db-go-tweets`
- configura plugin de autenticacao
- define usuario, senha e banco
- expoe a porta `3306`
- persiste dados em volume

### Linhas 16-17

```yaml
volumes:
  mysql_data:
```

- cria um volume nomeado para persistencia

## Relacao Entre Esses Arquivos

- `docker-compose.yml` sobe o banco
- `config.go` le as variaveis
- `mysql.go` conecta no banco
- `main.go` monta a aplicacao inteira
- `go.mod` declara bibliotecas usadas
