package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

var secretKey = []byte("secret-key")

type UserClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func ValidateToken(token string) (claims *UserClaims, err error) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secretKey, nil
	})

	if err != nil {
		return &UserClaims{}, err
	}

	if !parsedToken.Valid {
		return &UserClaims{}, errors.New("parsed token is invalid")
	}

	err = parsedToken.Claims.Valid()
	if err != nil {
		return &UserClaims{}, err
	}

	claims, ok := parsedToken.Claims.(*UserClaims)
	if !ok {
		return &UserClaims{}, errors.New("failed to cast claims to UserClaims")
	}

	return claims, nil
}

func CreateToken(username, user_id string) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256,
		UserClaims{
			Username: username,
			StandardClaims: jwt.StandardClaims{
				Subject:   user_id,
				ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
				IssuedAt:  time.Now().Unix(),
				Issuer:    "paygo",
				Audience:  "paygo_users",
			},
		},
	)

	token, err := claims.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}
	return token, nil
}
