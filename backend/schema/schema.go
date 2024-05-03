package schema

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Role struct {
	ID    uint `gorm:"primaryKey;autoIncrement"`
	Name  string
	Users []User `gorm:"foreignKey:RoleID"`
}
type User struct {
	UserID           uint              `json:"user_id" gorm:"primaryKey;autoIncrement"`
	RoleID           uint              `json:"role_id"`
	FirstName        string            `json:"first_name" form:"FirstName"`
	LastName         string            `json:"last_name" form:"LastName"`
	Address          string            `json:"address" form:"Address"`
	Email            string            `json:"email" form:"Email"`
	PhoneNumber      string            `json:"phone_number" form:"PhoneNumber"`
	Password         string            `json:"password" form:"Password"`
	Balance          float64           `json:"balance" gorm:"default:10000"`
	IsActive         bool              `json:"is_active" gorm:"default:true"`
	Products         []Product         `json:"products" gorm:"foreignKey:UserID"`
	Orders           []Order           `json:"orders" gorm:"foreignKey:UserID"`
	TransactionSends []TransactionHash `json:"transaction_sended" gorm:"foreignKey:UserID"`
	ImageID          uint              `json:"-" gorm:"foreignKey:ImageID"`
	Image            Image             `json:"image"`
}
type TransactionHash struct {
	TransactionID   string    `gorm:"primaryKey" json:"transaction_hash"`
	UserID          uint      `gorm:"foreignKey:UserID" json:"user_id"`
	TransactionDate time.Time `gorm:"type:timestamp" json:"transaction_date"`
}
type Category struct {
	CategoryID   uint `gorm:"primaryKey;autoIncrement"`
	CategoryName string
	IsActive     bool      `gorm:"default:true"`
	Products     []Product `gorm:"foreignKey:CategoryID"`
}
type Status struct {
	StatusID uint `gorm:"primaryKey;autoIncrement"`
	Status   string
	Products []Product `gorm:"foreignKey:StatusID"`
}
type Product struct {
	ProductID   uint           `json:"product_id" form:"product_id" gorm:"primaryKey;autoIncrement"`
	UserID      uint           `json:"user_id" form:"user_id"`
	CategoryID  uint           `json:"category_id" form:"category_id"`
	StatusID    uint           `json:"status_id" gorm:"default:2"`
	ProductName string         `json:"product_name" form:"product_name"`
	Description string         `json:"description" form:"description"`
	Price       float64        `json:"price" form:"price"`
	Quantity    uint           `json:"quantity" form:"quantity"`
	Images      []ProductImage `gorm:"foreignKey:ProductID"`
	Orders      []Order        `gorm:"foreignKey:ProductID"`
}
type ProductImage struct {
	ProductImageID uint `gorm:"primaryKey;autoIncrement"`
	ProductID      uint
	ImageID        uint `gorm:"foreignKey:ImageID"`
	Image          Image
}
type Image struct {
	ImageID   uint   `gorm:"primaryKey;autoIncrement" json:"image_id"`
	BucketKey string `json:"bucket_key"`
	Path      string `json:"path"`
}
type Order struct {
	OrderID     uint `gorm:"primaryKey;autoIncrement"`
	UserID      uint
	ProductID   uint
	OrderTotal  float64
	OrderDate   time.Time `gorm:"type:timestamp"`
	OrderStatus string
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
func (u *User) ChangePassword(db *gorm.DB, newPassword, currentPassword string) error {
	hashNewPass, err := HashPassword(newPassword)
	fmt.Println("user: ", u)
	if err != nil {
		return fmt.Errorf("hashing password failed")
	}
	match := CheckPasswordHash(currentPassword, u.Password)
	if !match {
		return fmt.Errorf("current password not match")
	}
	// Cập nhật mật khẩu người dùng trong cơ sở dữ liệu
	result := db.Model(u).Update("password", hashNewPass)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
func Migration() {
	dsn := "host=localhost user=postgres password=chicuong dbname=fitness-api port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	//db.Migrator().DropTable(&Role{}, &User{}, &Category{}, &Product{}, &Image{}, &Order{})
	db.AutoMigrate(&Role{}, &User{}, &Category{}, &Product{}, &Image{}, &Order{})
}
