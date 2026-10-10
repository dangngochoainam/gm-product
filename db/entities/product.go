package entities

type Product struct {
	ID          string  `gorm:"column:id;primaryKey;default:gen_random_uuid()"`
	Name        string  `gorm:"column:name;not null"`
	Code        string  `gorm:"column:code;not null;uniqueIndex"`
	Price       float64 `gorm:"column:price;not null;default:0"`
	Image       *string `gorm:"column:image"`
	Description string  `gorm:"column:description;not null"`

	BaseModel
}

func (Product) TableName() string {
	return "products"
}
