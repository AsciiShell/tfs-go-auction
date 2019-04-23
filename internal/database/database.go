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
	Debug       bool
	Migrate     bool
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
	db.LogMode(credential.Debug)
	result := DataBase{DB: db}
	if credential.Migrate {
		result.Migrate()
		logger.Info("Migrate completed")
	}
	return &result, nil
}
func (d *DataBase) constraintExists(table string, constraint string) bool {
	return d.DB.Exec(`SELECT 1 FROM pg_catalog.pg_constraint con
         INNER JOIN pg_catalog.pg_class rel ON rel.oid = con.conrelid
WHERE rel.relname = ? AND con.conname = ?;`, table, constraint).RowsAffected == 1
}
func (d *DataBase) Migrate() {
	d.DB.AutoMigrate(&user.User{}, &session.Session{}, &lot.Lot{})
	d.DB.Model(&session.Session{}).AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE")
	if d.DB.Exec("SELECT 1 FROM pg_type WHERE typname = 'lot_status'").RowsAffected == 0 {
		d.DB.Exec("CREATE TYPE lot_status  AS enum('created','active','finished')")
	}
	type constrain struct {
		Table string
		Name  string
		Rule  string
	}
	constraints := []constrain{
		{Table: "lots", Name: "lots_check_min_price", Rule: "CHECK(min_price >= 1)"},
		{Table: "lots", Name: "lots_check_price_step", Rule: "CHECK(price_step >= 1)"},
		{Table: "lots", Name: "lots_check_buy_price", Rule: "CHECK(buy_price >= min_price)"},
		{Table: "lots", Name: "lots_check_buyer_owner", Rule: "CHECK(creator_id != buyer_id)"},
		{Table: "lots", Name: "lots_check_end", Rule: "CHECK(end_at >= created_at OR end_at IS NULL)"},
	}
	for _, v := range constraints {
		if !d.constraintExists(v.Table, v.Name) {
			d.DB.Exec(fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT  %s %s", v.Table, v.Name, v.Rule))
		}
	}

	d.DB.Model(&lot.Lot{}).AddForeignKey("creator_id", "users(id)", "CASCADE", "CASCADE")
	d.DB.Model(&lot.Lot{}).AddForeignKey("buyer_id", "users(id)", "CASCADE", "CASCADE")
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
	if err := d.DB.First(s).Error; err != nil {
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

func (d *DataBase) attachUsersToLot(l *lot.Lot) {
	var write user.User
	d.DB.Where("id = ?", l.CreatorID).First(&write)
	write.IsShort = true
	l.Creator = &write
	if l.BuyerID != nil {
		var write2 user.User
		d.DB.Where("id = ?", *l.BuyerID).First(&write2)
		write2.IsShort = true
		l.Buyer = &write2
	}

}
func (d *DataBase) GetLots(condition lot.Lot) ([]lot.Lot, error) {
	var result []lot.Lot
	d.DB.Where(condition).Find(&result)
	for i := range result {
		d.attachUsersToLot(&result[i])
	}
	return result, nil
}

func (d *DataBase) GetLot(l *lot.Lot) error {
	if err := d.DB.Where(&l).First(&l).Error; err != nil {
		return errors.Wrapf(err, "lot not found %+v", l)
	}
	d.attachUsersToLot(l)
	return nil
}

func (d *DataBase) AddLot(l *lot.Lot) error {
	if err := d.DB.Create(&l).Error; err != nil {
		return errors.Wrap(err, "can't create lot")
	}
	d.attachUsersToLot(l)
	return nil
}

func (d *DataBase) UpdateUser(u *user.User, n *user.User) error {
	if err := d.DB.Where(&u).Update(n).Error; err != nil {
		return errors.Wrap(err, "can't update user")
	}
	return nil
}

func (d *DataBase) UpdateLot(n *lot.Lot) error {
	if err := d.DB.Model(&lot.Lot{}).Updates(*n).Error; err != nil {
		return errors.Wrap(err, "can't update lot")
	}
	if err := d.DB.Where(" id = ?", n.ID).First(&n).Error; err != nil {
		return errors.Wrapf(err, "lot not found %+v", n)
	}
	d.attachUsersToLot(n)
	return nil
}

func (d *DataBase) DeleteLot(l *lot.Lot) error {
	request := d.DB.Where(&l).Delete(&lot.Lot{})
	if request.Error != nil {
		return errors.Wrap(request.Error, "can't delete lot")
	}
	if request.RowsAffected == 0 {
		return fmt.Errorf("lot not found")
	}
	return nil
}
func (d *DataBase) GetOwnLots(l *lot.Lot, r *lot.Lot) ([]lot.Lot, error) {
	var result []lot.Lot
	d.DB.Where(l).Or(r).Find(&result)
	for i := range result {
		d.attachUsersToLot(&result[i])
	}
	return result, nil
}
func (d *DataBase) BuyLot(id int, owner int, price int) (lot.Lot, error) {
	tx := d.DB.Begin()
	defer tx.Commit()
	result := d.DB.Exec(`UPDATE lots
SET buy_price = ?,
    buyer_id = ?
WHERE id = ?
  AND deleted_at IS NULL
  AND status = 'active'
  AND creator_id != ?
  AND (buyer_id != ? OR buyer_id IS NULL)
  AND (buy_price < ? OR buy_price IS NULL)
  AND (? - min_price) % price_step = 0`, price, owner, id, owner, owner, price, price)
	if result.Error != nil {
		return lot.Lot{}, fmt.Errorf("can't buy lot :%+v", result.Error)
	}
	if result.RowsAffected == 0 {
		return lot.Lot{}, fmt.Errorf("can't buy lot, check lot status, owner and buyer statuses, your price: it should be more than a last price and equal n*step + start_price")

	}
	var lotResult lot.Lot
	if err := d.DB.Where("id = ?", id).First(&lotResult).Error; err != nil {
		tx.Rollback()
		return lot.Lot{}, errors.Wrapf(err, "can't fetch new lot, rollback")

	}
	d.attachUsersToLot(&lotResult)
	return lotResult, nil
}
