package user

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int       `gorm:"PRIMARY_KEY;AUTO_INCREMENT"`
	FirstName string    `gorm:"NOT NULL"`
	LastName  string    `gorm:"NOT NULL"`
	Birthday  time.Time `gorm:"type:date"`
	Email     string    `gorm:"NOT NULL;unique_index"`
	Password  string    `gorm:"NOT NULL"`
	CreatedAt time.Time `gorm:"NOT NULL"`
	UpdatedAt time.Time `gorm:"NOT NULL"`
	IsShort   bool
}
type userShort struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
type userFull struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Birthday  string    `json:"birthday"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

const BirthdayFormat = "2006-01-02"

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

func (u User) MarshalJSON() ([]byte, error) {
	if u.IsShort {
		return json.Marshal(userShort{ID: u.ID, FirstName: u.FirstName, LastName: u.LastName})
	}
	return json.Marshal(userFull{ID: u.ID, FirstName: u.FirstName, LastName: u.LastName, Birthday: u.Birthday.Format(BirthdayFormat), Email: u.Email, CreatedAt: u.CreatedAt})
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
