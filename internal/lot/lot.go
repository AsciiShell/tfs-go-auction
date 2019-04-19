package lot

import "time"

type Lot struct {
	ID          int       `json:"id" gorm:"PRIMARY_KEY;AUTO_INCREMENT"`
	CreatorID   int       `json:"creator_id" gorm:"NOT NULL"`
	Title       string    `json:"title" gorm:"NOT NULL"`
	Description string    `json:"description"`
	MinPrice    float64   `json:"min_price" gorm:"NOT NULL;type:numeric CHECK(min_price >= 1)"`
	PriceStep   float64   `json:"price_step" gorm:"NOT NULL;type:numeric CHECK(price_step >= 1);default:1"`
	CreatedAt   time.Time `json:"created_at" gorm:"NOT NULL"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"NOT NULL"`
	DeletedAT   time.Time `json:"deleted_at"`
}
