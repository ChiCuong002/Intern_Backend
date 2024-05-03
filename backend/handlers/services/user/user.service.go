package service

import (
	"fmt"
	storage "main/database"
	"main/helper/scope"
	helper "main/helper/struct"
	"main/models"
	"main/schema"
	"gorm.io/gorm"
)

const ZERO_VALUE_INT = 0

func UpdateUser(tx *gorm.DB, userData *helper.UserInsert) error {
	fmt.Println("service")
	updateData := make(map[string]interface{})

	if userData.FirstName != "" {
		updateData["first_name"] = userData.FirstName
	}
	if userData.LastName != "" {
		updateData["last_name"] = userData.LastName
	}
	if userData.Address != "" {
		updateData["address"] = userData.Address
	}
	if userData.Email != "" {
		updateData["email"] = userData.Email
	}
	if userData.Image != ZERO_VALUE_INT {
		updateData["image_id"] = userData.Image
	}

	fmt.Println("updateData: ", updateData)
	result := tx.Model(&schema.User{}).Where("user_id = ?", userData.UserID).Updates(updateData)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

const (
	ADMIN = 1
	USER  = 2
)

func Search(scope *gorm.DB, search string) *gorm.DB {
	if search != "" {
		scope = scope.Where("first_name LIKE ? OR last_name LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	return scope
}
func GetAllUserPagination(pagination helper.Pagination) (*helper.Pagination, error) {
	var db *gorm.DB = storage.GetDB()
	var users []models.User
	query := db.Model(&users)
	query = Search(query, pagination.Search)
	query = query.Scopes(scope.Paginate(query, &pagination))
	query.Find(&users)
	if query.RowsAffected == 0 {
		pagination.TotalPages = 1
	}
	pagination.Rows = users
	return &pagination, nil
}
func UserDetail(id uint) (helper.UserResponse, error) {
	var db *gorm.DB = storage.GetDB()
	fmt.Println("id: ", id)
	var user helper.UserResponse
	result := db.Model(&user).Preload("Image").First(&user, "user_id = ?", id)
	if result.Error != nil {
		//SELECT * FROM users WHERE id = 10;
		return user, fmt.Errorf("Failed to get user")
	}
	return user, nil
}
func BlockUser(id uint) (helper.UserResponse, error) {
	var db *gorm.DB = storage.GetDB()
	user, err := UserDetail(id)
	if err != nil {
		return user, err
	}
	db.Model(&user).Where("user_id = ?", id).Update("is_active", !user.IsActive)
	return user, nil
}
func AddBalanceByID(req helper.AddBalanceReq) error {
	db := storage.GetDB()
	user := schema.User{}
	result := db.Model(&user).First(&user, "user_id = ?", req.UserID)
	if result.Error != nil {
		return fmt.Errorf(result.Error.Error())
	}
	//update balance
	balance := user.Balance + req.AmountToAdd
	//update in db
	result = db.Model(&user).Where("user_id = ?", user.UserID).Update("balance", balance)
	if result.Error != nil {
		return fmt.Errorf(result.Error.Error())
	}
	return nil
}
func AddBalanceAllUser(tx *gorm.DB, req helper.Gif) error {
	users := []schema.User{}
	result := tx.Find(&users)
	if result.Error != nil {
		return fmt.Errorf(result.Error.Error())
	}
	for _, user := range users {
		// Update balance
		balance := user.Balance + req.Amount
		// Update in db
		result = tx.Model(&user).Where("user_id = ?", user.UserID).Update("balance", balance)
		if result.Error != nil {
			// Rollback transaction if there is an error
			tx.Rollback()
			return fmt.Errorf(result.Error.Error())
		}
	}

	// Commit transaction
	tx.Commit()

	return nil
}
func InsertTransactionHash(req schema.TransactionHash) error {
	db := storage.GetDB()
	result := db.Create(&req)
	if result.Error != nil {
		return fmt.Errorf(result.Error.Error())
	}
	return nil
}