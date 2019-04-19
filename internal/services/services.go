package services

import (
	"fmt"
	"time"

	"gitlab.com/asciishell/tfs-go-auction/internal/errs"

	"gitlab.com/asciishell/tfs-go-auction/internal/storage"

	"gitlab.com/asciishell/tfs-go-auction/internal/user"
	"golang.org/x/crypto/bcrypt"

	"github.com/pkg/errors"
	"gitlab.com/asciishell/tfs-go-auction/internal/session"
)

func Registry(u *user.User, storage *storage.Storage) error {
	if u.ID != 0 {
		return fmt.Errorf("user seems have been alredy registered %v", u)
	}
	if u.Email == "" || u.Password == "" {
		return errs.ErrEmptyCredits
	}
	hash, err := user.HashPassword(u.Password)
	if err != nil {
		return errors.Wrapf(err, "can't hash password %s", u.Password)
	}

	u.Password = hash
	if err = (*storage).AddUser(u); err != nil {
		return errors.Wrapf(err, "can't registry user")
	}
	return nil
}

func FindUserByEmail(email string, password string, storage *storage.Storage) (*user.User, error) {
	u := user.User{Email: email}
	if err := (*storage).GetUser(&u); err != nil {
		return nil, fmt.Errorf("user not found %s", email)
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil {
		return &u, nil
	}
	return nil, fmt.Errorf("wrong password for %s", email)
}

func FindUserByID(id int, storage *storage.Storage) (*user.User, error) {
	u := user.User{ID: id}
	if err := (*storage).GetUser(&u); err != nil {
		return nil, fmt.Errorf("user not found %d", id)
	}
	return &u, nil
}

func NewSession(userID int, storage *storage.Storage) (session.Session, error) {

	token, err := session.GenerateToken()
	if err != nil {
		return session.Session{}, errors.Wrapf(err, "can't generate token")
	}
	result := session.Session{SessionID: token, UserID: userID, CreatedAt: time.Now(), ValidUntil: time.Now().Add(session.TokenLifeTime)}
	if err = (*storage).AddSession(&result); err != nil {
		return session.Session{}, errors.Wrapf(err, "can't add session to database")
	}
	return result, nil
}

func GetSession(sessionID string, storage *storage.Storage) (*session.Session, error) {
	sess := session.Session{SessionID: sessionID}
	if err := (*storage).GetSession(&sess); err != nil {
		return nil, fmt.Errorf("session not found %s", sessionID)

	}
	return &sess, nil
}
