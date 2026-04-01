package auth

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AddressClaims struct {
	MailboxID uint   `json:"mailboxId"`
	Address   string `json:"address"`
	jwt.RegisteredClaims
}

type AddressJWTManager struct {
	mu          sync.RWMutex
	secret      []byte
	expireHours int
}

func NewAddressJWTManager(secret string, expireHours int) *AddressJWTManager {
	return &AddressJWTManager{secret: []byte(secret), expireHours: expireHours}
}

func (m *AddressJWTManager) Generate(mailboxID uint, address string) (string, error) {
	m.mu.RLock()
	secret := append([]byte(nil), m.secret...)
	expireHours := m.expireHours
	m.mu.RUnlock()

	now := time.Now()
	exp := now.Add(time.Duration(expireHours) * time.Hour)
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
	return t.SignedString(secret)
}

func (m *AddressJWTManager) Parse(token string) (*AddressClaims, error) {
	m.mu.RLock()
	secret := append([]byte(nil), m.secret...)
	m.mu.RUnlock()

	parsed, err := jwt.ParseWithClaims(token, &AddressClaims{}, func(t *jwt.Token) (any, error) {
		return secret, nil
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

func (m *AddressJWTManager) Update(secret string, expireHours int) {
	m.mu.Lock()
	m.secret = []byte(secret)
	m.expireHours = expireHours
	m.mu.Unlock()
}
