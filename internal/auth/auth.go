package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"gitlab.com/asciishell/tfs-go-auktion/internal/database"

	"github.com/pkg/errors"
	"gitlab.com/asciishell/tfs-go-auktion/internal/services"
	"gitlab.com/asciishell/tfs-go-auktion/internal/session"
)

func Signin(email string, password string, storage database.Storage) (session.Session, error) {
	u, err := services.FindUserByEmail(email, password, nil)
	if err != nil {
		return session.Session{}, err
	}
	sess, err := services.NewSession(u.ID, storage)
	if err != nil {
		return session.Session{}, errors.Wrapf(err, "can't create session for user ID %d", u.ID)
	}
	return sess, nil
}

func preValidateToken(r *http.Request, storage database.Storage) (*session.Session, error) {
	token := strings.Split(r.Header.Get("Authorization"), " ")
	if len(token) != 2 || token[0] != "Bearer" {
		return nil, fmt.Errorf("invalid token %s", r.Header.Get("Authorization"))
	}
	return services.GetSession(token[1], storage)
}
func ValidateToken(r *http.Request, storage database.Storage) bool {
	sess, ok := preValidateToken(r, storage)
	return ok == nil && sess.ValidUntil.After(time.Now())
}
func ValidateTokenUser(r *http.Request, uID int, storage database.Storage) bool {
	sess, ok := preValidateToken(r, storage)
	return ok == nil && sess.ValidUntil.After(time.Now()) && sess.UserID == uID
}
