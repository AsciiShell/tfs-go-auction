package session

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/pkg/errors"
)

type Session struct {
	SessionID  string    `json:"session_id"`
	UserID     int       `json:"user_id"`
	CreatedAt  time.Time `json:"created_at"`
	ValidUntil time.Time `json:"valid_until"`
}

var sessions []Session

const TokenLifeTime = time.Hour * 24

const alphabet = "qwertyuiopasdfghjlzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890"

const tokenLen = 255

func generateToken() (string, error) {
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

func NewSession(userID int) (Session, error) {
	token, err := generateToken()
	if err != nil {
		return Session{}, errors.Wrapf(err, "can't generate token")
	}
	result := Session{SessionID: token, UserID: userID, CreatedAt: time.Now(), ValidUntil: time.Now().Add(TokenLifeTime)}
	sessions = append(sessions, result)
	return result, nil
}

func GetSession(sessionID string) (*Session, error) {
	for i, v := range sessions {
		if sessionID == v.SessionID {
			return &sessions[i], nil
		}
	}
	return nil, fmt.Errorf("session not found %s", sessionID)
}

func (s Session) Marshal() ([]byte, error) {
	return json.Marshal(s)
}
