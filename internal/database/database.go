package database

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"gitlab.com/asciishell/tfs-go-auction/internal/lot"
	"gitlab.com/asciishell/tfs-go-auction/internal/session"
	"gitlab.com/asciishell/tfs-go-auction/internal/user"
	"gitlab.com/asciishell/tfs-go-auction/pkg/log"

	// Registry postgres
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pkg/errors"
)

type DataBase struct {
	DB *gorm.DB
}

type DBCredential struct {
	User        string
	Password    string
	Host        string
	Database    string
	Repetitions int
}

func NewDataBaseStorage(credential DBCredential) (*DataBase, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable&fallback_application_name=fintech-app", credential.User, credential.Password, credential.Host, credential.Database)
	var err error
	var db *gorm.DB
	logger := log.New()
	for i := 0; i < credential.Repetitions; i++ {
		logger.Infof("Take %d/%d to connect to database", i+1, credential.Repetitions)
		db, err = gorm.Open("postgres", dsn)
		if err == nil && db.DB().Ping() == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "can't connect to database, dsn %s", dsn)
	}
	return &DataBase{DB: db}, nil
}

func (d *DataBase) Migrate() {
	d.DB.AutoMigrate(&user.User{}, &session.Session{}, &lot.Lot{})
	d.DB.Model(&session.Session{}).AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE")
}
func (d *DataBase) GetUser(u *user.User) error {
	if err := d.DB.Where(&u).First(&u).Error; err != nil {
		return errors.Wrapf(err, "user not found %+v", u)
	}
	return nil
}
func (d *DataBase) AddUser(u *user.User) error {
	if err := d.DB.Create(&u).Error; err != nil {
		return errors.Wrap(err, "can't create user")
	}
	return nil
}

func (d *DataBase) GetSession(s *session.Session) error {
	if err := d.DB.Where(&s).First(&s).Error; err != nil {
		return errors.Wrapf(err, "session not found %+v", s)
	}
	return nil
}

func (d *DataBase) AddSession(s *session.Session) error {
	if err := d.DB.Create(&s).Error; err != nil {
		return errors.Wrap(err, "can't create session")
	}
	return nil
}

func (d *DataBase) GetLots() ([]lot.Lot, error) {
	var result []lot.Lot
	d.DB.Find(&result)
	return result, nil
}

func (d *DataBase) GetLot(l *lot.Lot) error {
	if err := d.DB.Where(&l).First(&l).Error; err != nil {
		return errors.Wrapf(err, "lot not found %+v", l)
	}
	return nil
}

func (d *DataBase) AddLot(l *lot.Lot) error {
	if err := d.DB.Create(&l).Error; err != nil {
		return errors.Wrap(err, "can't create lot")
	}
	return nil
}

func (d *DataBase) SetLot(l *lot.Lot) error {
	if err := d.DB.Save(&l).Error; err != nil {
		return errors.Wrap(err, "can't set lot")
	}
	return nil
}
func (d *DataBase) UpdateUser(u *user.User, n *user.User) error {
	if err := d.DB.Where(&u).Update(n).Error; err != nil {
		return errors.Wrap(err, "can't update user")
	}
	return nil
}
