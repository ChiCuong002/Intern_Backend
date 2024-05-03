package controllers

import (
	//"crypto/ecdsa"
	"fmt"
	storage "main/database"
	services "main/handlers/services/normal"
	paginationHelper "main/helper/struct"
	"strconv"

	//"main/helper/validation"
	"main/schema"
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

const (
	INFINITY      = 1000000
	LIMIT_DEFAULT = 10
	PAGE_DEFAULT  = 1
	SORT_DEFAULT  = " product_id desc"
	FILTER_DEFAULT = 0
)

func sortString(sort string) string {
	order := sort[0]
	sortString := sort[1:]
	fmt.Println("sortString: ", sortString)
	fmt.Println("ASCII: ", int(order))
	fmt.Println("order: ", order)
	fmt.Println("rune(order): ", rune(order))
	fmt.Println("rune('+'): ", rune('+'))
	fmt.Println("rune('-'): ", rune('-'))

	if rune(order) == '+' || rune(order) == ' ' {
		sortString = sortString + " asc"
	} else if rune(order) == '-' {
		sortString = sortString + " desc"
	} else {
		sortString = ""
	}
	fmt.Println("sortString: ", sortString)
	return sortString
}

type LoginRequest struct {
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

func Login(c echo.Context) error {
	var LoginRequest LoginRequest
	if err := c.Bind(&LoginRequest); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid JSON format"})
	}

	//user := schema.User{}
	username := LoginRequest.PhoneNumber
	password := LoginRequest.Password
	fmt.Println("Received phone_number:", username)
	fmt.Println("Received password:", password)
	//get user token error
	user, token, err := services.Login(username, password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"user":    user,
		"token":   token,
		"message": "User login successfull",
	})
}
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
func RegisterUser(c echo.Context) error {
	// Đọc dữ liệu từ request
	db := storage.GetDB()
	var newUser schema.User
	err := c.Bind(&newUser)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "Invalid request payload",
		})
	}
	// Kiểm tra xem email đã tồn tại chưa
	var existingUser schema.User
	result := db.Where("phone_number = ?", newUser.PhoneNumber).First(&existingUser)
	if len(newUser.PhoneNumber) != 10 {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "Phone number is invalid",
		})
	}
	if result.RowsAffected > 0 {
		return c.JSON(http.StatusConflict, echo.Map{
			"message": "Phone number is already exists",
		})
	}
	if len(newUser.Password) < 8 {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "Password should have at least 8 characters",
		})
	}
	hash, err := HashPassword(newUser.Password)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "Hash password failed",
		})
	}
	match := CheckPasswordHash(newUser.Password, hash)
	if !match {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "Compare password failed",
		})
	}
	newUser.Password = hash
	if newUser.RoleID == 0 {
		newUser.RoleID = 2
	}
	//default image
	newUser.ImageID = 1
	//
	result = db.Create(&newUser)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": result.Error.Error(),
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "Register successfull",
	})
}
func CategoriesDropDown(c echo.Context) error {
	categories, err := services.CategoriesDropDown()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"message":    "Get all categories successfull",
		"categories": categories,
	})
}
func HomePage(c echo.Context) error {
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil {
		page = PAGE_DEFAULT
	}
	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil {
		limit = LIMIT_DEFAULT
	}
	sort := c.QueryParam("sort")
	if sort != "" {
		sort = sortString(sort)
	} else {
		sort = SORT_DEFAULT
	}
	search := c.QueryParam("search")
	filter, err := strconv.Atoi(c.QueryParam("filter"))
	if err != nil {
		filter = FILTER_DEFAULT
	}
	pagination := paginationHelper.Pagination{
		Page:   page,
		Limit:  limit,
		Sort:   sort,
		Search: search,
		Filter: filter,
	}

	products, err := services.GetHomePageProduct(pagination)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, products)
}
