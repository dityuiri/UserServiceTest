package handler

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func (s *Server) generateJWTToken(id string) (string, error) {
	// By default, set the expiration time for 2 minutes
	expirationTime := time.Now().Add(2 * time.Minute)
	claims := &jwt.MapClaims{
		"id":  id,
		"exp": jwt.NewNumericDate(expirationTime),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.JWTSecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *Server) retrieveAndGetIdFromJWTToken(ctx echo.Context) (string, error) {
	token, err := s.retrieveJWTToken(ctx)
	if err != nil {
		return "", err
	}

	// Get ID from JWT token
	userId, err := s.getIdFromJWTToken(token)
	if err != nil {
		return "", err
	}

	return userId, nil
}

func (s *Server) retrieveJWTToken(c echo.Context) (string, error) {
	authHeader := c.Request().Header.Get("Authorization")

	// Get the auth header
	if authHeader == "" {
		return "", errors.New("missing JWT token")
	}

	// Check if it's valid token by escape the "Bearer"
	token := s.extractToken(authHeader)
	if token == "" {
		return "", errors.New("invalid JWT token")
	}

	return token, nil
}

func (s *Server) extractToken(authHeader string) string {
	// Check if the header has the "Bearer" scheme
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}

func (s *Server) getIdFromJWTToken(token string) (string, error) {
	claims := &jwt.MapClaims{}
	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.JWTSecretKey), nil
	})

	if err != nil {
		return "", err
	}

	if !tkn.Valid {
		return "", errors.New("token is no longer valid")
	}

	validClaims := tkn.Claims.(*jwt.MapClaims)
	return (*validClaims)["id"].(string), nil
}
