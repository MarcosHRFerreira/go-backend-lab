// Package jwt provides JWT creation and validation helpers.
package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func CreateToken(id int, username, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       id,
		"username": username,
		"exp":      time.Now().Add(60 * time.Minute).Unix(),
	},
	)

	key := []byte(secretKey)
	tokenStr, err := token.SignedString(key)

	return tokenStr, err
}

func ValidateToken(tokenStr, secretKey string, withClaimValidate bool) (int, string, error) {
	var (
		key    = []byte(secretKey)
		claims = jwt.MapClaims{}
		token  *jwt.Token
		err    error
	)

	if withClaimValidate {
		token, err = jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return key, nil
		})
	} else {
		token, err = jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return key, nil
		}, jwt.WithoutClaimsValidation())

	}
	if err != nil {
		return 0, "", err
	}

	if !token.Valid {
		return 0, "", errors.New("token is not valid")
	}

	// jwt.MapClaims devolve numeros como float64 apos o parse,
	// entao o id nao pode ser convertido direto para int.
	idFloat, ok := claims["id"].(float64)
	if !ok {
		return 0, "", errors.New("invalid token id")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return 0, "", errors.New("invalid token username")
	}

	return int(idFloat), username, nil

}

func ValidadeToken(tokenStr, secretKey string, withClaimValidate bool) (int, string, error) {
	return ValidateToken(tokenStr, secretKey, withClaimValidate)
}
