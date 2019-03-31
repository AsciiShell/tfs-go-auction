package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gitlab.com/asciishell/tfs-go-auktion/internal/session"
	"gitlab.com/asciishell/tfs-go-auktion/internal/user"
)

func Signin(email string, password string) (session.Session, error) {
	u, err := user.FindUserByEmail(email, password)
	if err != nil {
		return session.Session{}, err
	}
	sess, err := session.NewSession(u.ID)
	if err != nil {
		return session.Session{}, errors.Wrapf(err, "can't create session for user ID %d", u.ID)
	}
	return sess, nil
}

func preValidateToken(r *http.Request) (*session.Session, error) {
	token := strings.Split(r.Header.Get("Authorization"), " ")
	if len(token) != 2 || token[0] != "Bearer" {
		return nil, fmt.Errorf("invalid token %s", r.Header.Get("Authorization"))
	}
	return session.GetSession(token[1])
}
func ValidateToken(r *http.Request) bool {
	sess, ok := preValidateToken(r)
	return ok == nil && sess.ValidUntil.After(time.Now())
}
func ValidateTokenUser(r *http.Request, uID int) bool {
	sess, ok := preValidateToken(r)
	return ok == nil && sess.ValidUntil.After(time.Now()) && sess.UserID == uID
}
