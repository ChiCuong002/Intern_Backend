package services

import (
	"errors"
	"fmt"
	storage "main/database"
	"main/helper/scope"
	helper "main/helper/struct"
	paginationHelper "main/helper/struct"
	productHelper "main/helper/struct/product"
	"main/schema"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Login(username, password string) (schema.User, string, error) {
	var user schema.User
	db := storage.GetDB()
	//check phone number in db
	result := db.Where("phone_number = ?", username).Preload("Image").First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return schema.User{}, "", fmt.Errorf("user not registered: %v", result.Error)
	}
	//check if user is active
	if !user.IsActive {
		return schema.User{}, "", fmt.Errorf("user is blocked")
	}
	//check match password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return schema.User{}, "", fmt.Errorf("password is not match")
	}
	//create jwt token
	claims := &helper.JwtCustomClaims{
		UserId:    user.UserID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		IsAdmin:   user.RoleID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return schema.User{}, "", fmt.Errorf("signed token error")
	}
	return user, t, nil
}
func CategoriesDropDown() ([]helper.CategoriesDropDown, error) {
	db := storage.GetDB()
	category := []helper.CategoriesDropDown{}
	result := db.Select("category_id, category_name").Find(&category, schema.Category{IsActive: true})
	if result.Error != nil {
		return category, result.Error
	}
	if result.RowsAffected == 0 {
		return category, fmt.Errorf("can't found any categories")
	}
	return category, nil
}
func SearchProducts(query *gorm.DB, search string) *gorm.DB {
	if search != "" {
		query = query.Where("product_name like ?", "%"+search+"%")
	}
	return query

}
func GetHomePageProduct(pagination paginationHelper.Pagination) (*paginationHelper.Pagination, error) {
	db := storage.GetDB()
	products := []productHelper.DetailProductRes{}
	query := db.Model(&products).Where("status_id = 1")
	if pagination.Filter != 0 {
		query = query.Where("category_id = ?", pagination.Filter)
	}
	query = query.Preload("User").Preload("User.Image").Preload("Category").Preload("Status").Preload("ProductImages.Image")
	query = SearchProducts(query, pagination.Search)
	query = query.Scopes(scope.Paginate(query, &pagination))
	query.Find(&products)
	pagination.Rows = products
	return &pagination, nil
}
