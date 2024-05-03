package controllers

import (
	"fmt"
	storage "main/database"
	services "main/handlers/services/category"
	helper "main/helper/struct"
	"main/schema"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

const (
	LIMIT_DEFAULT = 10
	PAGE_DEFAULT  = 1
	SORT_DEFAULT  = " category_id desc"
)

func sortString(sort string) string {
	order := sort[0]
	sortString := sort[1:]
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
func GetCategories(c echo.Context) error {
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
	pagination := helper.Pagination{
		Page:   page,
		Limit:  limit,
		Sort:   sort,
		Search: search,
	}
	categories, err := services.GetAllCategories(pagination)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, categories)
}
func DetailCategory(c echo.Context) error {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}
	fmt.Println("idInt: ", idInt)
	category, err := services.GetDetailCategory(uint(idInt))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err,
		})
	}
	return c.JSON(http.StatusOK, category)
}

type AddCategoryRequest struct {
	NameCategory string `json:"category_name"`
}

func AddCategory(c echo.Context) error {
	var addCategoryRequest AddCategoryRequest
	if err := c.Bind(&addCategoryRequest); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Server error",
		})
	}
	
	db := storage.GetDB()
	newCategory := schema.Category{
		CategoryName: addCategoryRequest.NameCategory,
	}

	result := db.Create(&newCategory)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": result.Error.Error(),
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"category": newCategory,
		"message":  "Thêm loại sản phẩm thành công!",
	})
}
func EditCategory(c echo.Context) error {
	var requestData struct {
		IDCategory      string `json:"category_id"`
		NewNameCategory string `json:"category_name"`
		IsActive        bool   `json:"is_active"`
	}
	db := storage.GetDB()
	err := c.Bind(&requestData)
	if err != nil {
		//fmt.Println("Error binding request:", err)
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "lấy loại không được",
			"error":   err.Error(),
		})
	}

	var categoyToUpdate schema.Category
	result := db.First(&categoyToUpdate, requestData.IDCategory)
	if result.Error != nil {
		return c.JSON(http.StatusNotFound, echo.Map{
			"message": "Không tìm thấy loại sản phẩm",
		})
	}

	err = services.ChangeNameCategory(&categoyToUpdate, db, requestData.NewNameCategory)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Lỗi thay đổi loại: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"category": requestData,
		"message":  "Đổi tên loại thành công",
	})
}
