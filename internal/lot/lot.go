package lot

import (
	"fmt"
	"strings"
	"time"

	"gitlab.com/asciishell/tfs-go-auction/internal/user"
)

type Status int

const (
	Created Status = iota
	Active
	Finished
)

func (s Status) String() string {
	switch s {
	case Created:
		return "created"
	case Active:
		return "active"
	case Finished:
		return "finished"
	default:
		return ""
	}
}

func NewStatus(s string) (Status, error) {
	switch strings.ToLower(s) {
	case "created":
		return Created, nil
	case "active":
		return Active, nil
	case "finished":
		return Finished, nil
	default:
		return 0, fmt.Errorf("can't recognise string %s", s)
	}
}

type Lot struct {
	ID          int        `json:"id" gorm:"PRIMARY_KEY;AUTO_INCREMENT"`
	Title       string     `json:"title" gorm:"NOT NULL"`
	Description *string    `json:"description"`
	MinPrice    float64    `json:"min_price" gorm:"NOT NULL;type:numeric"`
	PriceStep   float64    `json:"price_step" gorm:"NOT NULL;type:numeric;default:1"`
	BuyPrice    *float64   `json:"buy_price,omitempty" gorm:"type:numeric"`
	Status      string     `json:"status" gorm:"NOT NULL;type:lot_status;default:'created'"`
	EndAt       time.Time  `json:"end_at" gorm:"NOT NULL"`
	CreatorID   int        `json:"-" gorm:"NOT NULL"`
	Creator     *user.User `json:"creator" gorm:"-"`
	BuyerID     *int       `json:"-" gorm:""`
	Buyer       *user.User `json:"buyer,omitempty" gorm:"-"`
	CreatedAt   time.Time  `json:"created_at" gorm:"NOT NULL"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"NOT NULL"`
	DeletedAt   *time.Time `json:"-"`
}
