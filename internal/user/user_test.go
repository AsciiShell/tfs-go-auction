package user

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.com/asciishell/tfs-go-auction/internal/session"
	"golang.org/x/crypto/bcrypt"
)

func TestUser_Update(t *testing.T) {
	r := require.New(t)
	newYear := time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local)
	birthday := time.Now()
	u1 := User{ID: 1, Password: "123", Email: "test@golang.org", FirstName: "FirstName", LastName: "LastName", Birthday: newYear}
	u2 := User{ID: 2, Password: "new2", Email: "new2@golang.org", FirstName: "New2", LastName: "New2"}
	u3 := User{ID: 3, Password: "new3", Email: "new3@golang.org", FirstName: "New3", LastName: "New3", Birthday: birthday}
	u1.Update(u2)
	r.Equal(u1.FirstName, u2.FirstName)
	r.Equal(u1.LastName, u2.LastName)
	r.Equal(u1.Birthday, newYear)
	r.Equal(u1.ID, 1)
	r.Equal(u1.Password, "123")
	r.Equal(u1.Email, "test@golang.org")
	u1.Update(u3)
	r.Equal(u1.FirstName, u3.FirstName)
	r.Equal(u1.LastName, u3.LastName)
	r.Equal(u1.Birthday, birthday)
	r.Equal(u1.ID, 1)
	r.Equal(u1.Password, "123")
	r.Equal(u1.Email, "test@golang.org")
}

func TestHashPassword(t *testing.T) {
	r := require.New(t)
	password, err := session.GenerateToken()
	r.NoError(err)
	hash, err := HashPassword(password)
	r.NoError(err)
	r.Nil(bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)))
}

func TestUser_UnmarshalJSON(t *testing.T) {
	r := require.New(t)
	reader := bytes.NewReader([]byte(`{
  "first_name": "Павел",
  "last_name": "Дуров",
  "birthday": "1984-10-10",
  "email": "durov@telegram.org",
  "password": "qwerty"
}`))
	var u User
	err := json.NewDecoder(reader).Decode(&u)
	r.NoError(err)
	r.Equal(u.ID, 0)
	r.Equal(u.FirstName, "Павел")
	r.Equal(u.LastName, "Дуров")
	r.Equal(u.Birthday, time.Date(1984, 10, 10, 0, 0, 0, 0, time.UTC))
	r.Equal(u.Email, "durov@telegram.org")
	r.Equal(u.Password, "qwerty")
}
