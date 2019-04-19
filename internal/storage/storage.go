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
	GetLots() ([]lot.Lot, error)
	GetLot(l *lot.Lot) error
	AddLot(l *lot.Lot) error
	SetLot(l *lot.Lot) error
}
