package user

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gitlab.com/asciishell/tfs-go-auktion/pkg/date"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Birthday  date.Date `json:"birthday"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var users []*User

func lastID() int {
	result := 0
	for _, v := range users {
		if v.ID > result {
			result = v.ID
		}
	}
	return result + 1
}

func hashPassword(password string) (string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.Wrapf(err, "can't hash password %s", password)
	}
	return string(passwordHash), nil
}
func (u *User) Registry() error {
	if u.ID != 0 {
		return fmt.Errorf("user seems have been alredy registered %v", u)
	}
	hash, err := hashPassword(u.Password)
	if err != nil {
		return err
	}
	u.ID = lastID()
	u.Password = hash
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	users = append(users, u)
	return nil
}

func (u *User) Update(new User) {
	if new.FirstName != "" {
		u.FirstName = new.FirstName
	}
	if new.LastName != "" {
		u.LastName = new.LastName
	}
	if !new.Birthday.EqualTime(time.Time{}) {
		u.Birthday = new.Birthday
	}
	u.UpdatedAt = time.Now()
}

func FindUserByEmail(email string, password string) (*User, error) {
	for i, v := range users {
		if v.Email == email && bcrypt.CompareHashAndPassword([]byte(v.Password), []byte(password)) == nil {
			return users[i], nil
		}

	}
	return nil, fmt.Errorf("user not found %s_%s", email, password)
}

func FindUserByID(id int) (*User, error) {
	for i, v := range users {
		if v.ID == id {
			return users[i], nil
		}
	}
	return nil, fmt.Errorf("user not found %d", id)
}
