package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var TokenTime = map[string]time.Duration{
	"active_token":  time.Minute * 30,
	"refresh_token": time.Hour * 24 * 7,
}

type Authenticator interface {
	GenerateToken(id int32, tokenTime string) (string, error)
	ValidateToken(token string) (*jwt.Token, error)
	ValidateRefreshToken(token string) (*jwt.Token, error)
}

type AuthJwt struct {
	secret, aud, iss string
}

func NewJwt(secret, aud, iss string) *AuthJwt {
	return &AuthJwt{}
}

func (a *AuthJwt) GenerateToken(id int32, tokenTime string) (string, error) {
	claims := jwt.MapClaims{
		"sub": id,
		"exp": time.Now().Add(TokenTime[tokenTime]).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": a.iss,
		"aud": a.aud,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(a.secret))
	if err != nil {
		return "", fmt.Errorf("failed generate token :%v", err)
	}

	return tokenString, nil
}

func (a *AuthJwt) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return []byte(a.secret), nil
	})
}

func (a *AuthJwt) ValidateRefreshToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return []byte(a.secret), nil
	},
		jwt.WithoutClaimsValidation(),
	)
}
