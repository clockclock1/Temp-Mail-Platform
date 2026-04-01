package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AddressClaims struct {
	MailboxID uint   `json:"mailboxId"`
	Address   string `json:"address"`
	jwt.RegisteredClaims
}

type AddressJWTManager struct {
	secret      []byte
	expireHours int
}

func NewAddressJWTManager(secret string, expireHours int) *AddressJWTManager {
	return &AddressJWTManager{secret: []byte(secret), expireHours: expireHours}
}

func (m *AddressJWTManager) Generate(mailboxID uint, address string) (string, error) {
	now := time.Now()
	exp := now.Add(time.Duration(m.expireHours) * time.Hour)
	claims := AddressClaims{
		MailboxID: mailboxID,
		Address:   address,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			Subject:   fmt.Sprintf("mailbox:%d", mailboxID),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(m.secret)
}

func (m *AddressJWTManager) Parse(token string) (*AddressClaims, error) {
	parsed, err := jwt.ParseWithClaims(token, &AddressClaims{}, func(t *jwt.Token) (any, error) {
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*AddressClaims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
