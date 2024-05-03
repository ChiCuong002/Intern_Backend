package helper

type DetailProductRes struct {
	ProductID     uint           `json:"product_id" gorm:"primaryKey;autoIncrement"`
	UserID        uint           `json:"-" gorm:"foreignKey:UserID"`
	User          User           `json:"user"`
	CategoryID    uint           `json:"-" gorm:"foreignKey:CategoryID"`
	Category      Category       `json:"category"`
	ProductName   string         `json:"product_name"`
	Description   string         `json:"description"`
	Price         float64        `json:"price"`
	Quantity      uint           `json:"quantity"`
	StatusID      uint           `json:"-" gorm:"foreignKey:StatusID"`
	Status        Status         `json:"status"`
	ProductImages []ProductImage `json:"product_images" gorm:"foreignKey:ProductID"`
}
type Status struct {
	StatusID uint   `json:"status_id" gorm:"primaryKey;autoIncrement"`
	Status   string `json:"status_name"`
}
type ProductImage struct {
	ProductImageID uint  `json:"product_image_id" gorm:"primaryKey;autoIncrement"`
	ProductID      uint  `json:"product_id"`
	ImageID        uint  `json:"image_id" gorm:"foreignKey:ImageID"`
	Image          Image `json:"image"`
}

type Image struct {
	ImageID   uint   `json:"image_id" gorm:"primaryKey;autoIncrement"`
	BucketKey string `json:"bucket_key"`
	Path      string `json:"path"`
}
type User struct {
	UserID    uint   `json:"user_id" gorm:"primaryKey;autoIncrement"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	ImageID   uint   `json:"-" gorm:"foreignKey:ImageID"`
	Image     Image  `json:"image"`
	PhoneNumber string `json:"phone_number"`
}
type Category struct {
	CategoryID   uint   `json:"category_id" gorm:"primaryKey;autoIncrement"`
	CategoryName string `json:"category_name"`
}

func (DetailProductRes) TableName() string {
	return "products"
}

func (ProductImage) TableName() string {
	return "product_images"
}
func (Image) TableName() string {
	return "images"
}
func (User) TableName() string {
	return "users"
}
