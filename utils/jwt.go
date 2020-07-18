package utils

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
)

var engine = struct {
	secret []byte
}{}

func EncodeJwtToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(engine.secret)

	return tokenString, err
}

func DecodeJwtToken(tokenString string, claims jwt.Claims) (jwt.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return engine.secret, nil
	})

	if err != nil {
		return nil, err
	}

	return token.Claims, nil
}

func InitJwtEngine(secret string) {
	if secret == "" {
		logrus.Fatal("Cat't get JWT Secret, please specifiy it in config")
	}

	engine.secret = []byte(secret)
}
