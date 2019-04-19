package errs

import "errors"

var ErrUnauthorized = errors.New("неавторизованный запрос")
var ErrNotFound = errors.New("контент по переданному идентификатору не найден")
var ErrWrongData = errors.New("неверные входные данные")
var ErrEmptyCredits = errors.New("email and password should not be blank")
