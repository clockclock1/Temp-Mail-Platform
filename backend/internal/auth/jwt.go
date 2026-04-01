package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint     `json:"userId"`
	Username string   `json:"username"`
	Role     string   `json:"role"`
	Perms    []string `json:"perms"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret      []byte
	expireHours int
}

func NewJWTManager(secret string, expireHours int) *JWTManager {
	return &JWTManager{secret: []byte(secret), expireHours: expireHours}
}

func (j *JWTManager) GenerateToken(userID uint, username, role string, perms []string) (string, error) {
	now := time.Now()
	exp := now.Add(time.Duration(j.expireHours) * time.Hour)
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		Perms:    perms,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTManager) ParseToken(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		return j.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
