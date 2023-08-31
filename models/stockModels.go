package models

import (
	"encoding/json"

	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"
)

type StockData struct {
	gorm.Model
	Ticker     string
	OpenPrice  float64
	ClosePrice float64
	High       float64
	Low        float64
	Volume     int
}

type TransactionType string

const (
	Buy  TransactionType = "buy"
	Sell TransactionType = "sell"
)

type Transaction struct {
	gorm.Model
	Ticker            string
	TransactionType   TransactionType
	TransactionVolume int
	TransactionPrice  float64
	UserID            uint
	User              User `gorm:"foreignKey:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type User struct {
	gorm.Model
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"unique"`
	Email    string
	Balance  float64
	Password string
}

func (u User) MarshalBinary() ([]byte, error) {
	serializedUser, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}
	return serializedUser, nil
}

func (u *User) UnmarshalBinary(data []byte) error {
	err := json.Unmarshal(data, u)
	if err != nil {
		return err
	}
	return nil
}

type JWTClaims struct {
	jwt.StandardClaims
	UserID uint `json:"user_id"`
}
