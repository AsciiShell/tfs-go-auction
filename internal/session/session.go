package session

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	"time"

	"github.com/pkg/errors"
)

type Session struct {
	SessionID  string    `json:"session_id" gorm:"PRIMARY_KEY"`
	UserID     int       `json:"user_id" gorm:"NOT NULL"`
	CreatedAt  time.Time `json:"created_at" gorm:"NOT NULL"`
	ValidUntil time.Time `json:"valid_until" gorm:"NOT NULL"`
}

const TokenLifeTime = time.Hour * 24

const alphabet = "qwertyuiopasdfghjlzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890"

const tokenLen = 255

func GenerateToken() (string, error) {
	result := make([]uint8, tokenLen)
	for i := 0; i < tokenLen; i++ {
		res, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", errors.Wrapf(err, "can't generate token")
		}
		result[i] = alphabet[res.Int64()]
	}
	return string(result), nil
}

func (s Session) Marshal() ([]byte, error) {
	return json.Marshal(s)
}
