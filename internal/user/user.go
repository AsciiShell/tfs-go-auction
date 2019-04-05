package user

import (
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int       `json:"id" gorm:"PRIMARY_KEY;AUTO_INCREMENT"`
	FirstName string    `json:"first_name" gorm:"NOT NULL"`
	LastName  string    `json:"last_name" gorm:"NOT NULL"`
	Birthday  time.Time `json:"birthday" gorm:"type:date"`
	Email     string    `json:"email" gorm:"NOT NULL;unique_index"`
	Password  string    `json:"password" gorm:"NOT NULL"`
	CreatedAt time.Time `json:"created_at" gorm:"NOT NULL"`
	UpdatedAt time.Time `json:"updated_at" gorm:"NOT NULL"`
}

func HashPassword(password string) (string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.Wrapf(err, "can't hash password %s", password)
	}
	return string(passwordHash), nil
}

func (u *User) Update(new User) {
	if new.FirstName != "" {
		u.FirstName = new.FirstName
	}
	if new.LastName != "" {
		u.LastName = new.LastName
	}
	if !new.Birthday.Equal(time.Time{}) {
		u.Birthday = new.Birthday
	}
	u.UpdatedAt = time.Now()
}
