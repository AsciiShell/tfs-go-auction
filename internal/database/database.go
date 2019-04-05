package database

import (
	"fmt"

	"gitlab.com/asciishell/tfs-go-auktion/internal/session"
	"gitlab.com/asciishell/tfs-go-auktion/internal/user"

	"github.com/jinzhu/gorm"
	// Registry postgres
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pkg/errors"
)

type Storage interface {
	Migrate()
	GetUser(u *user.User) error
	AddUser(u *user.User) error
	GetSession(s *session.Session) error
	AddSession(s *session.Session) error
}

type DataBase struct {
	DB *gorm.DB
}

type DBCredential struct {
	User     string
	Password string
	Host     string
	Table    string
}

func NewDataBaseStorage(credential DBCredential) (*DataBase, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable&fallback_application_name=fintech-app", credential.User, credential.Password, credential.Host, credential.Table)
	db, err := gorm.Open("postgres", dsn)
	if err != nil {
		return nil, errors.Wrapf(err, "can't connect to database, dsn %s", dsn)
	}
	err = db.DB().Ping()
	if err != nil {
		return nil, errors.Wrap(err, "can't ping database")
	}
	return &DataBase{DB: db}, nil
}

func (d *DataBase) Migrate() {
	d.DB.AutoMigrate(&user.User{}, &session.Session{})
	d.DB.Model(&session.Session{}).AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE")
}
func (d *DataBase) GetUser(u *user.User) error {
	d.DB = d.DB.Where(u).First(u)
	if err := d.DB.Error; err != nil {
		return errors.Wrapf(err, "user not found %+v", u)
	}
	return nil
}
func (d *DataBase) AddUser(u *user.User) error {
	d.DB = d.DB.Create(u)
	if err := d.DB.Error; err != nil {
		return errors.Wrap(err, "can't create user")
	}
	return nil
}

func (d *DataBase) GetSession(s *session.Session) error {
	d.DB = d.DB.Where(s).First(s)
	if err := d.DB.Error; err != nil {
		return errors.Wrapf(err, "session not found %+v", s)
	}
	return nil
}

func (d *DataBase) AddSession(s *session.Session) error {
	d.DB = d.DB.Create(s)
	if err := d.DB.Error; err != nil {
		return errors.Wrap(err, "can't create session")
	}
	return nil
}
