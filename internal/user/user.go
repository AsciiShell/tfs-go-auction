package user

import (
	"encoding/json"
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
	Password  string    `json:"-" gorm:"NOT NULL"`
	CreatedAt time.Time `json:"created_at" gorm:"NOT NULL"`
	UpdatedAt time.Time `json:"-" gorm:"NOT NULL"`
}

func (u *User) UnmarshalJSON(b []byte) error {
	var user map[string]string
	err := json.Unmarshal(b, &user)
	if err != nil {
		return errors.Wrapf(err, "can't unmarshal json")
	}
	for key, value := range user {
		switch key {
		case "first_name":
			u.FirstName = value
		case "last_name":
			u.LastName = value
		case "birthday":
			t, err := time.Parse("2006-01-02", value)
			if err != nil {
				return errors.Wrapf(err, "can't parse date %s", value)
			}
			u.Birthday = t
		case "email":
			u.Email = value
		case "password":
			u.Password = value
		}
	}
	return nil
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
