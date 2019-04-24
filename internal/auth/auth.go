package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gitlab.com/asciishell/tfs-go-auction/internal/errs"
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
	var token string
	headerPair := strings.Split(r.Header.Get("Authorization"), " ")
	if len(headerPair) == 2 && headerPair[0] == "Bearer" {
		token = headerPair[1]
	}
	cookie, err := r.Cookie("BearerToken")
	if token == "" && err == nil {
		token = cookie.Value
	}
	sess, err := services.GetSession(token, s)
	if err != nil || sess.ValidUntil.Before(time.Now()) {
		return nil, errs.ErrNotFound
	}
	return sess, nil
}
