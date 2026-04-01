package models

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	DisplayName  string    `gorm:"size:128" json:"displayName"`
	Active       bool      `gorm:"default:true" json:"active"`
	RoleID       uint      `gorm:"not null" json:"roleId"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Role struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Name        string       `gorm:"size:64;uniqueIndex;not null" json:"name"`
	Description string       `gorm:"size:255" json:"description"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

type Permission struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Key         string    `gorm:"size:64;uniqueIndex;not null" json:"key"`
	Description string    `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Domain struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:255;uniqueIndex;not null" json:"name"`
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	CreatedBy uint      `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Mailbox struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	LocalPart    string     `gorm:"size:128;not null;index:idx_mailbox_unique,unique" json:"localPart"`
	DomainID     uint       `gorm:"not null;index:idx_mailbox_unique,unique" json:"domainId"`
	Domain       Domain     `json:"domain"`
	OwnerID      uint       `gorm:"not null;index" json:"ownerId"`
	Owner        User       `json:"owner"`
	Description  string     `gorm:"size:255" json:"description"`
	Enabled      bool       `gorm:"default:true" json:"enabled"`
	ExpiresAt    *time.Time `gorm:"index" json:"expiresAt"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	LastReceived *time.Time `json:"lastReceived"`
}

type Message struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	MailboxID  uint      `gorm:"not null;index" json:"mailboxId"`
	Mailbox    Mailbox   `json:"mailbox"`
	MessageID  string    `gorm:"size:255;index" json:"messageId"`
	FromAddr   string    `gorm:"size:512" json:"from"`
	ToAddr     string    `gorm:"size:512" json:"to"`
	Subject    string    `gorm:"size:1024" json:"subject"`
	RawPath    string    `gorm:"size:1024" json:"rawPath"`
	TextBody   string    `gorm:"type:text" json:"textBody"`
	HTMLBody   string    `gorm:"type:text" json:"htmlBody"`
	Size       int64     `json:"size"`
	ReceivedAt time.Time `gorm:"index" json:"receivedAt"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (m Mailbox) Address() string {
	return m.LocalPart + "@" + m.Domain.Name
}
