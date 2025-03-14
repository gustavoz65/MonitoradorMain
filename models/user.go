package models

import (
	"sync"
	"time"

	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Username string  `gorm:"unique;not null"`
	Email    string  `gorm:"unique;not null"`
	Password string  `gorm:"not null"`
	Emails   []Email `gorm:"foreignkey:UserID"`
}

type Email struct {
	gorm.Model
	Address string `gorm:"unique;not null"`
	UserID  uint
}

var (
	Mutex            sync.Mutex
	LogFileName      = "log.txt"
	EmailNotificacao string
	Delay            = 5
	IntervaloLimpeza = 7 * 24 * time.Hour
	Ticker           *time.Ticker
)
