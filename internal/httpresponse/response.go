// Package httpresponse provides shared HTTP response helpers for Gin handlers.
package httpresponse

import (
	"errors"
	"fmt"
	"go-tweets/internal/apperror"
	"go-tweets/internal/observability/logctx"
	"log/slog"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Message string        `json:"message"`
	Errors  []ErrorDetail `json:"errors,omitempty"`
}

func JSONError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorResponse{
		Message: message,
	})
}

func JSONErrorFromErr(c *gin.Context, statusCode int, err error) {
	JSONError(c, statusCode, err.Error())
}

func JSONAppError(c *gin.Context, err error) {
	statusCode := apperror.StatusCode(err)
	if statusCode >= http.StatusInternalServerError {
		requestLogger := logctx.FromContext(c.Request.Context())
		args := []any{
			slog.Int("status_code", statusCode),
			slog.String("error", err.Error()),
		}

		if cause := errors.Unwrap(err); cause != nil {
			args = append(args, slog.String("cause", cause.Error()))
		}

		requestLogger.Error("request failed", args...)
	}

	JSONError(c, statusCode, err.Error())
}

func AbortJSONError(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, ErrorResponse{
		Message: message,
	})
}

func BindAndValidateJSON(c *gin.Context, validate *validator.Validate, req interface{}) bool {
	// Reject malformed JSON before running struct validation rules.
	// Rejeita JSON malformado antes de executar as regras de validacao do struct.
	if err := c.ShouldBindJSON(req); err != nil {
		JSONError(c, http.StatusBadRequest, "invalid request body")
		return false
	}

	// Run declarative validation tags so handlers stay focused on orchestration.
	// Executa as tags declarativas de validacao para que os handlers continuem focados na orquestracao.
	if err := validate.Struct(req); err != nil {
		JSONValidationError(c, err, req)
		return false
	}

	return true
}

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

		// Return JSON field names instead of Go struct names to keep the API contract user-facing.
		// Retorna os nomes JSON dos campos em vez dos nomes do struct Go para manter o contrato da API orientado ao cliente.
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

func ParseIntParam(c *gin.Context, name string) (int, bool) {
	value := c.Param(name)
	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		JSONError(c, http.StatusBadRequest, fmt.Sprintf("%s must be a valid integer", name))
		return 0, false
	}

	return parsedValue, true
}

func ParseMinInt64Query(c *gin.Context, name string, defaultValue string, min int64) (int64, bool) {
	rawValue := c.DefaultQuery(name, defaultValue)
	parsedValue, err := strconv.ParseInt(rawValue, 10, 64)
	if err != nil || parsedValue < min {
		JSONError(c, http.StatusBadRequest, fmt.Sprintf("%s must be a valid integer greater than or equal to %d", name, min))
		return 0, false
	}

	return parsedValue, true
}

func jsonFieldNames(req interface{}) map[string]string {
	reqType := reflect.TypeOf(req)
	if reqType == nil {
		return map[string]string{}
	}

	// Accept both struct values and pointers so handlers can pass request DTOs naturally.
	// Aceita tanto valores struct quanto ponteiros para que os handlers possam passar DTOs de requisicao de forma natural.
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
