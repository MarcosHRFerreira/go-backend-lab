# `response.go`: Leitura Linha por Linha

## Como Usar Este Arquivo

Este guia foi feito para voce estudar o arquivo de apoio HTTP do projeto com calma, quase como uma aula comentada.

Arquivo foco:

- [response.go](file:///c:/Users/marco/OneDrive/Projects/go-tweets/internal/httpresponse/response.go)

## Visao Geral

O arquivo `internal/httpresponse/response.go` centraliza comportamentos comuns da camada HTTP:

- resposta JSON de erro
- resposta padronizada para erros de validacao
- bind e validacao de body JSON
- parse de parametros de rota
- parse de query params numericos

Na pratica, ele existe para deixar os handlers menores, mais legiveis e menos repetitivos.

## Linhas 1-2

```go
// Package httpresponse provides shared HTTP response helpers for Gin handlers.
package httpresponse
```

### O que essas linhas fazem

- a primeira linha e um comentario de pacote
- a segunda linha declara que o arquivo pertence ao pacote `httpresponse`

### Por que isso importa

- em Go, comentarios de pacote ajudam documentacao automatica
- o nome do pacote indica a responsabilidade do arquivo
- aqui a ideia e clara: utilitarios de resposta HTTP

## Linhas 4-14: imports

```go
import (
	"fmt"
	"go-tweets/internal/apperror"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)
```

### O que estudar aqui

- `fmt`: formatacao de strings
- `apperror`: padronizacao de erros de aplicacao
- `net/http`: constantes de status HTTP
- `reflect`: inspecao de tipos em tempo de execucao
- `strconv`: conversao de string para inteiro
- `strings`: utilitarios de texto
- `gin`: framework HTTP usado no projeto
- `validator`: biblioteca de validacao por tags

### Leitura arquitetural

Esse import block mostra que o arquivo atua como uma ponte entre:

- o framework HTTP
- a validacao de entrada
- a padronizacao de erros

## Linhas 16-19: `ErrorDetail`

```go
type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
```

### O que a estrutura representa

- um erro individual de validacao
- `Field`: qual campo falhou
- `Message`: qual foi a falha

### Exemplo mental

Se o usuario enviar um email invalido, a API pode responder algo como:

```json
{
  "field": "email",
  "message": "must be a valid email"
}
```

## Linhas 21-24: `ErrorResponse`

```go
type ErrorResponse struct {
	Message string        `json:"message"`
	Errors  []ErrorDetail `json:"errors,omitempty"`
}
```

### O que essa estrutura representa

- o formato padrao de resposta de erro da API

### Campos

- `Message`: mensagem geral do erro
- `Errors`: lista opcional de detalhes

### O que `omitempty` quer dizer

- se `Errors` estiver vazio, esse campo nao aparece no JSON

### Resultado pratico

A mesma estrutura atende:

- erro simples
- erro detalhado de validacao

## Linhas 26-30: `JSONError`

```go
func JSONError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorResponse{
		Message: message,
	})
}
```

### O que a funcao faz

- recebe o contexto HTTP
- recebe um status code
- recebe uma mensagem
- devolve uma resposta JSON padronizada

### Por que isso existe

Sem esse helper, cada handler repetiria o mesmo bloco de resposta.

## Linhas 32-34: `JSONErrorFromErr`

```go
func JSONErrorFromErr(c *gin.Context, statusCode int, err error) {
	JSONError(c, statusCode, err.Error())
}
```

### O que a funcao faz

- recebe um `error`
- converte para string com `err.Error()`
- reaproveita `JSONError`

### Ganho de design

- evita repeticao
- concentra o formato de resposta em um unico lugar

## Linhas 36-38: `JSONAppError`

```go
func JSONAppError(c *gin.Context, err error) {
	JSONError(c, apperror.StatusCode(err), err.Error())
}
```

### O que essa funcao faz

- recebe um erro vindo da camada de aplicacao
- extrai o status HTTP apropriado
- responde com a mensagem do erro

### Papel arquitetural

Essa funcao liga:

- camada de negocio
- camada HTTP

Sem obrigar o handler a decidir manualmente todo status code.

## Linhas 40-44: `AbortJSONError`

```go
func AbortJSONError(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, ErrorResponse{
		Message: message,
	})
}
```

### Diferenca para `JSONError`

- `JSONError` apenas responde
- `AbortJSONError` responde e interrompe a cadeia do Gin

### Quando isso e util

- middleware
- autenticacao
- validacao de acesso

## Linhas 46-58: `BindAndValidateJSON`

```go
func BindAndValidateJSON(c *gin.Context, validate *validator.Validate, req interface{}) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		JSONError(c, http.StatusBadRequest, "invalid request body")
		return false
	}

	if err := validate.Struct(req); err != nil {
		JSONValidationError(c, err, req)
		return false
	}

	return true
}
```

### O que essa funcao concentra

- leitura do body JSON
- validacao do struct recebido

### Linha conceitual 1

```go
if err := c.ShouldBindJSON(req); err != nil {
```

- tenta ler o JSON da requisicao e preencher `req`
- se o body estiver malformado, entra no bloco de erro

### Linha conceitual 2

```go
JSONError(c, http.StatusBadRequest, "invalid request body")
```

- devolve `400 Bad Request`
- informa que o corpo da requisicao e invalido

### Linha conceitual 3

```go
if err := validate.Struct(req); err != nil {
```

- usa a biblioteca validator para aplicar as tags do struct
- exemplo: `required`, `email`, `min`, `eqfield`

### Linha conceitual 4

```go
JSONValidationError(c, err, req)
```

- delega a montagem da resposta detalhada de validacao

### Resultado final

- retorna `true` quando tudo esta correto
- retorna `false` quando ja respondeu erro ao cliente

## Linhas 60-85: `JSONValidationError`

```go
func JSONValidationError(c *gin.Context, err error, req interface{}) {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		JSONError(c, http.StatusBadRequest, "validation failed")
		return
	}

	fieldNames := jsonFieldNames(req)
	details := make([]ErrorDetail, 0, len(validationErrors))
	for _, validationErr := range validationErrors {
		fieldName := fieldNames[validationErr.Field()]
		if fieldName == "" {
			fieldName = strings.ToLower(validationErr.Field())
		}

		details = append(details, ErrorDetail{
			Field:   fieldName,
			Message: validationMessage(validationErr, fieldNames),
		})
	}

	c.JSON(http.StatusBadRequest, ErrorResponse{
		Message: "validation failed",
		Errors:  details,
	})
}
```

### Primeira etapa

```go
validationErrors, ok := err.(validator.ValidationErrors)
```

- tenta converter o erro generico para o tipo especifico da biblioteca validator

### Por que isso e necessario

- o `error` generico nao carrega todos os detalhes diretamente
- o tipo `validator.ValidationErrors` permite acessar:
  - campo que falhou
  - tag que falhou
  - parametro da validacao

### Segunda etapa

```go
fieldNames := jsonFieldNames(req)
```

- cria um mapa entre nome do campo Go e nome do campo JSON

Exemplo:

- `PasswordConfirm` -> `password_confirm`

### Terceira etapa

```go
details := make([]ErrorDetail, 0, len(validationErrors))
```

- cria um slice vazio para acumular detalhes
- ja reserva capacidade suficiente

### Quarta etapa

```go
for _, validationErr := range validationErrors {
```

- percorre todos os erros de validacao encontrados

### Quinta etapa

```go
fieldName := fieldNames[validationErr.Field()]
```

- tenta descobrir o nome JSON correto do campo

### Sexta etapa

```go
if fieldName == "" {
	fieldName = strings.ToLower(validationErr.Field())
}
```

- fallback defensivo
- se nao houver tag JSON, usa o nome do campo em minusculo

### Setima etapa

```go
details = append(details, ErrorDetail{
	Field:   fieldName,
	Message: validationMessage(validationErr, fieldNames),
})
```

- monta um item de erro detalhado
- usa `validationMessage` para transformar a tag tecnica em mensagem humana

### O fechamento da funcao

- responde com `400`
- mensagem geral `validation failed`
- lista detalhada em `errors`

## Linhas 87-96: `ParseIntParam`

```go
func ParseIntParam(c *gin.Context, name string) (int, bool) {
	value := c.Param(name)
	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		JSONError(c, http.StatusBadRequest, fmt.Sprintf("%s must be a valid integer", name))
		return 0, false
	}

	return parsedValue, true
}
```

### O que essa funcao faz

- le um parametro da rota
- tenta converter para inteiro

### Exemplo de uso

Se a rota for:

```text
/tweets/:post_id
```

e o usuario mandar:

```text
/tweets/abc
```

- `strconv.Atoi("abc")` falha
- a API responde `400`

### Padrao de retorno

- retorna `(valor, true)` em caso de sucesso
- retorna `(0, false)` em caso de erro

## Linhas 98-107: `ParseMinInt64Query`

```go
func ParseMinInt64Query(c *gin.Context, name string, defaultValue string, min int64) (int64, bool) {
	rawValue := c.DefaultQuery(name, defaultValue)
	parsedValue, err := strconv.ParseInt(rawValue, 10, 64)
	if err != nil || parsedValue < min {
		JSONError(c, http.StatusBadRequest, fmt.Sprintf("%s must be a valid integer greater than or equal to %d", name, min))
		return 0, false
	}

	return parsedValue, true
}
```

### O que essa funcao faz

- le um query param
- usa valor padrao se ele nao vier
- converte para `int64`
- garante um valor minimo

### Exemplo mental

Para `page` e `limit`:

- `page=1` funciona
- `page=0` falha se `min` for `1`
- `page=abc` tambem falha

### Por que `int64`

- e uma escolha mais segura para valores numericos vindos da URL
- combina bem com paginação e limites

## Linhas 109-135: `jsonFieldNames`

```go
func jsonFieldNames(req interface{}) map[string]string {
	reqType := reflect.TypeOf(req)
	if reqType == nil {
		return map[string]string{}
	}

	if reqType.Kind() == reflect.Ptr {
		reqType = reqType.Elem()
	}

	if reqType.Kind() != reflect.Struct {
		return map[string]string{}
	}

	fieldNames := make(map[string]string, reqType.NumField())
	for i := 0; i < reqType.NumField(); i++ {
		field := reqType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		fieldNames[field.Name] = strings.Split(jsonTag, ",")[0]
	}

	return fieldNames
}
```

### O objetivo dessa funcao

- descobrir como os campos do struct aparecem no JSON

### Linha conceitual 1

```go
reqType := reflect.TypeOf(req)
```

- descobre o tipo real recebido

### Linha conceitual 2

```go
if reqType.Kind() == reflect.Ptr {
	reqType = reqType.Elem()
}
```

- se `req` for ponteiro, pega o tipo apontado

### Linha conceitual 3

```go
if reqType.Kind() != reflect.Struct {
	return map[string]string{}
}
```

- a funcao so faz sentido para structs

### Linha conceitual 4

```go
jsonTag := field.Tag.Get("json")
```

- le a tag JSON de cada campo

### Linha conceitual 5

```go
fieldNames[field.Name] = strings.Split(jsonTag, ",")[0]
```

- guarda o nome principal do campo JSON
- ignora opcoes extras como `omitempty`

### Exemplo pratico

Se o struct tiver:

```go
PasswordConfirm string `json:"password_confirm"`
```

o mapa final tera:

```text
PasswordConfirm -> password_confirm
```

## Linhas 137-155: `validationMessage`

```go
func validationMessage(fieldError validator.FieldError, fieldNames map[string]string) string {
	switch fieldError.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email"
	case "min":
		return fmt.Sprintf("must have at least %s characters", fieldError.Param())
	case "eqfield":
		targetField := fieldNames[fieldError.Param()]
		if targetField == "" {
			targetField = strings.ToLower(fieldError.Param())
		}

		return fmt.Sprintf("must match %s", targetField)
	default:
		return "is invalid"
	}
}
```

### O papel dessa funcao

- traduz regras de validacao em mensagens amigaveis

### Como ler o `switch`

- `required`: campo obrigatorio
- `email`: email invalido
- `min`: tamanho minimo nao atendido
- `eqfield`: um campo deve ser igual a outro

### O caso `eqfield`

Esse caso e importante para cenarios como:

- `password_confirm` deve ser igual a `password`

O codigo tenta descobrir o nome JSON do campo alvo para que a mensagem fique boa para quem consome a API.

### O `default`

- serve como fallback
- se surgir uma regra nao tratada, a API ainda responde algo coerente

## Sintese Final

Esse arquivo nao representa regra de negocio do dominio. Ele representa um **apoio de infraestrutura HTTP**.

Ele existe para:

- padronizar respostas
- reduzir repeticao nos handlers
- melhorar a clareza das mensagens de erro
- validar entradas de forma consistente
- simplificar parse de parametros e query params

## Mapa Mental Deste Arquivo

```text
handler
  -> bind do JSON
  -> validacao
  -> parse de params
  -> resposta padronizada de erro
```

## Frase Final Para Memorizar

O `response.go` funciona como uma camada auxiliar da interface HTTP, traduzindo entradas invalidadas e erros tecnicos em respostas JSON consistentes para toda a API.
