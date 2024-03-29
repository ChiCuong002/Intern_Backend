package services

import (
	"errors"
	"fmt"
	storage "main/database"
	helper "main/helper/struct"
	"main/schema"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Login(username, password string) (schema.User, string, error) {
	var user schema.User
	db := storage.GetDB()
	//check phone number in db
	result := db.Where("phone_number = ?", username).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return schema.User{}, "", fmt.Errorf("User not registered: %v", result.Error)
	}
	//check if user is active
	if !user.IsActive {
		return schema.User{}, "", fmt.Errorf("User is blocked")
	}
	//check match password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return schema.User{}, "", fmt.Errorf("Password is not match")
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
		return schema.User{}, "", fmt.Errorf("Signed token error")
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
		return category, fmt.Errorf("Can't found any categories")
	}
	return category, nil
}
