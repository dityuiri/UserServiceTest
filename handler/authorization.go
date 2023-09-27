package handler

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (s *Server) generateJWTToken(id string) (string, error) {
	// By default, set the expiration time for 3 minutes
	expirationTime := time.Now().Add(3 * time.Minute)
	claims := &jwt.MapClaims{
		"Id": id,
		"RegisteredClaims": jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.JWTSecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
