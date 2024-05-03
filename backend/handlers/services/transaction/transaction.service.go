package transaction

import (
	storage "main/database"
	"main/helper/scope"
	paginationHelper "main/helper/struct"
	"main/schema"

	"gorm.io/gorm"
)

func SearchProducts(query *gorm.DB, search string) *gorm.DB {
	if search != "" {
		query = query.Where("transaction_id like ?", "%"+search+"%")
	}
	return query

}
func GetAllTransactionHash(pagination paginationHelper.Pagination) (*paginationHelper.Pagination, error) {
	db := storage.GetDB()
	transactions := []schema.TransactionHash{}
	query := db.Model(&transactions)
	query = SearchProducts(query, pagination.Search)
	query = query.Scopes(scope.Paginate(query, &pagination))
	query.Find(&transactions)
	pagination.Rows = transactions
	return &pagination, nil
}
func GetUserTransactionHash(pagination paginationHelper.Pagination, userId uint) (*paginationHelper.Pagination, error) {
	db := storage.GetDB()
	transactions := []schema.TransactionHash{}
	query := db.Model(&transactions)
	query = SearchProducts(query, pagination.Search)
	query = query.Scopes(scope.Paginate(query, &pagination))
	query.Find(&transactions).Where("user_id = ?", userId)
	pagination.Rows = transactions
	return &pagination, nil
}
