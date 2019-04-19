package auth

import (
	"net/http"
	"strings"
	"time"

	"gitlab.com/asciishell/tfs-go-auction/internal/errs"

	"github.com/pkg/errors"
	"gitlab.com/asciishell/tfs-go-auction/internal/services"
	"gitlab.com/asciishell/tfs-go-auction/internal/session"
	"gitlab.com/asciishell/tfs-go-auction/internal/storage"
)

func Signin(email string, password string, storage *storage.Storage) (session.Session, error) {
	u, err := services.FindUserByEmail(email, password, storage)
	if err != nil {
		return session.Session{}, err
	}
	sess, err := services.NewSession(u.ID, storage)
	if err != nil {
		return session.Session{}, errors.Wrapf(err, "can't create session for user ID %d", u.ID)
	}
	return sess, nil
}

func HandleToken(r *http.Request, s *storage.Storage) (*session.Session, error) {
	token := strings.Split(r.Header.Get("Authorization"), " ")
	if len(token) != 2 || token[0] != "Bearer" {
		return nil, errs.ErrUnauthorized
	}
	sess, err := services.GetSession(token[1], s)
	if err != nil || sess.ValidUntil.Before(time.Now()) {
		return nil, errs.ErrNotFound
	}
	return sess, nil
}
