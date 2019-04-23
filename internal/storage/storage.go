package storage

import (
	"gitlab.com/asciishell/tfs-go-auction/internal/lot"
	"gitlab.com/asciishell/tfs-go-auction/internal/session"
	"gitlab.com/asciishell/tfs-go-auction/internal/user"
)

type Storage interface {
	Migrate()

	GetUser(u *user.User) error
	AddUser(u *user.User) error
	UpdateUser(u *user.User, n *user.User) error

	GetSession(s *session.Session) error
	AddSession(s *session.Session) error

	GetLots(condition lot.Lot) ([]lot.Lot, error)
	GetLot(l *lot.Lot) error
	GetOwnLots(l *lot.Lot, r *lot.Lot) ([]lot.Lot, error)
	BuyLot(id int, owner int, price int) (lot.Lot, error)
	AddLot(l *lot.Lot) error
	UpdateLot(n *lot.Lot) error
	DeleteLot(l *lot.Lot) error
	CloseLots() (int, error)
}
