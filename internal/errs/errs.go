package errs

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

type Err struct {
	Err string `json:"error"`
}

func (e Err) Error() string {
	return e.Err
}
func NewError(err error) Err {
	return Err{Err: err.Error()}
}
func NewErrorStr(err string, args ...interface{}) Err {
	return Err{Err: fmt.Sprintf(err, args...)}
}
func (e Err) StringJSON() string {
	result, _ := json.Marshal(e)
	return string(result)
}

var ErrUnauthorized = errors.New("неавторизованный запрос")
var ErrNotFound = errors.New("контент по переданному идентификатору не найден")
var ErrEmptyCredits = errors.New("email and password should not be blank")
