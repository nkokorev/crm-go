package models

import (
	"github.com/dgrijalva/jwt-go"
	"testing"
	"time"
)

func TestCreateToken(t *testing.T) {

	expiresAt := time.Now().Add(time.Minute * 10).Unix()
	claims := Token{
		12,
		77,
		jwt.StandardClaims{
			ExpiresAt: expiresAt,
			Issuer:    "AuthServer",
		},
	}
	_, err := claims.CreateAESToken()
	if err!= nil {
		t.Error("Cant create jwt token: ", err)
	}
}

