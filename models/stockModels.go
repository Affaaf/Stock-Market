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
type Transaction struct {
	gorm.Model
	Ticker            string
	TransactionType   string
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
	// Serialize the user data to bytes using JSON encoding
	serializedUser, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}
	return serializedUser, nil
}

// Implement BinaryUnmarshaler interface
func (u *User) UnmarshalBinary(data []byte) error {
	// Deserialize the bytes to the User struct using JSON decoding
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
