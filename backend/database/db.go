package storage

import (
	"fmt"
	"log"
	"main/schema"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai", dbHost, dbUser, dbPass, dbName, dbPort)

	db, err = gorm.Open(postgres.Open(connStr), &gorm.Config{})

	if err != nil {
		log.Fatal("failed to connect database")
	}

	fmt.Println("Successfully connected to database")
}
func GetDB() *gorm.DB {
	return db
}
func Migration() {
	db.AutoMigrate(&schema.Role{}, &schema.User{}, &schema.TransactionHash{}, &schema.Category{}, &schema.Status{}, &schema.Product{}, &schema.ProductImage{}, &schema.Image{}, &schema.Order{})
}
func DropColumns(tableName string, columnName string) {
	err := db.Migrator().DropColumn(tableName, columnName)
	if err != nil {
		log.Fatal("Failed to drop column: ", err)
	} else {
		fmt.Println("Successfully dropped column")
	}
}

func DropTable(tableName string) {
	err := db.Migrator().DropTable(tableName)
	if err != nil {
		log.Fatal("Failed to drop table: ", err)
	} else {
		fmt.Println("Successfully dropped table")
	}
}
