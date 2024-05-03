package controller

import (
	//"main/schema"
	"fmt"
	"log"
	storage "main/database"
	imageServices "main/handlers/services/image"
	service "main/handlers/services/user"
	userServices "main/handlers/services/user"
	"main/helper/aws"
	helper "main/helper/struct"
	imageHelper "main/helper/struct/product"
	"main/schema"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	LIMIT_DEFAULT       = 10
	PAGE_DEFAULT        = 1
	SORT_DEFAULT        = " user_id desc"
	DEFAULT_IMAGE_VALUE = 1
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
func GetAllUser(c echo.Context) error {
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
	users, err := userServices.GetAllUserPagination(pagination)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, users)
}
func DetailUser(c echo.Context) error {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}
	fmt.Println("id: ", idInt)
	user, err := userServices.UserDetail(uint(idInt))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	fmt.Println("user: ", user)
	return c.JSON(http.StatusOK, user)
}

func ChangePasswordUsers(c echo.Context) error {
	db := storage.GetDB()
	var requestData struct {
		UserID          uint   `json:"UserID"`
		CurrentPassword string `json:"CurrentPassword"`
		NewPassword     string `json:"NewPassword"`
	}
	err := c.Bind(&requestData)
	if err != nil {
		fmt.Println("Error binding request:", err)
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "Binding data failed",
		})
	}

	fmt.Println("User ID:", requestData.UserID)
	fmt.Println("New Password:", requestData.NewPassword)

	var userToUpdate schema.User
	result := db.First(&userToUpdate, requestData.UserID)
	if result.Error != nil {
		return c.JSON(http.StatusNotFound, echo.Map{
			"message": "User not found",
		})
	}

	// Thay đổi mật khẩu của người dùng
	err = userToUpdate.ChangePassword(db, requestData.NewPassword, requestData.CurrentPassword)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "Password is changed",
	})
}
func BlockUser(c echo.Context) error {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}
	fmt.Println("id: ", idInt)
	user, err := userServices.BlockUser(uint(idInt))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err,
		})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"message": "Block/Unblock user successfully",
		"user":    user,
	})
}

func UpdateUser(c echo.Context) error {
	fmt.Println("Updated user")
	var imageID uint
	user := helper.UserInsert{}
	userID := c.Get("userID").(uint)
	data := helper.UpdateData{}
	c.Bind(&data)
	// //Get multipart form from the request
	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid form data")
	}
	// Get the image files from the form
	image, ok := form.File["image"]
	//begin transaction
	tx := storage.GetDB().Begin()
	if ok && len(image) > 0 {
		userByID, err := service.UserDetail(userID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		//load aws service
		awsService, err := aws.ConfigureAWSService()
		// Generate a unique key for the S3 bucket
		bucketKey := uuid.New().String()
		// Open the image file
		src, err := image[0].Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
		defer src.Close()
		//create image path
		imagePath := aws.GetImagePath(bucketKey)
		//didn't have image (default image)
		if userByID.Image.ImageID == DEFAULT_IMAGE_VALUE {
			fmt.Println("default image")
			// Upload the image to AWS
			err = awsService.UploadFile(os.Getenv("BUCKETNAME"), bucketKey, src)
			if err != nil {
				log.Println("Failed to upload the image:", err)
				return c.JSON(http.StatusInternalServerError, echo.Map{
					"message": err.Error(),
				})
			} else {
				log.Println("Image uploaded successfully")
			}
			//insert the image into database
			img := &imageHelper.ImageInsert{
				BucketKey: bucketKey,
				Path:      imagePath,
			}
			err = imageServices.InsertImage(tx, img)
			if err != nil {
				tx.Rollback()
				return c.JSON(http.StatusInternalServerError, "Failed to insert image")
			}
			imageID = img.ImageID
			fmt.Println("img: ", img)
			fmt.Println("imageID: ", imageID)
			user = helper.UserInsert{
				UserID:    userID,
				FirstName: data.FirstName,
				LastName:  data.LastName,
				Address:   data.Address,
				Email:     data.Email,
				Image:     imageID,
			}
			fmt.Println("user: ", user)
			//already have an image
		} else if imageID := userByID.Image.ImageID; imageID != DEFAULT_IMAGE_VALUE {
			fmt.Println("have images")
			image, err := imageServices.GetImageByID(imageID)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, err.Error())
			}
			//update the image path with cache
			//image.Path = imagePath
			//update image in database
			err = imageServices.UpdateImage(tx, &image)
			if err != nil {
				tx.Rollback()
				return c.JSON(http.StatusInternalServerError, "Failed to update image")
			}
			//update image in cloud
			err = awsService.UpdateFile(os.Getenv("BUCKETNAME"), image.BucketKey, src)
			if err != nil {
				log.Println("Failed to upload the image:", err)
				return c.JSON(http.StatusInternalServerError, echo.Map{
					"message": err.Error(),
				})
			} else {
				log.Println("Image uploaded successfully")
			}
		}
	} else {
		user = helper.UserInsert{
			UserID:    userID,
			FirstName: data.FirstName,
			LastName:  data.LastName,
			Address:   data.Address,
			Email:     data.Email,
		}
	}
	err = service.UpdateUser(tx, &user)
	if err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}
	//commit transaction
	tx.Commit()
	return c.JSON(http.StatusOK, echo.Map{
		"message": "User infomation is updated successfully",
	})
}
func AddBalanceByID(c echo.Context) error {
	addBalanceReq := helper.AddBalanceReq{}
	if err := c.Bind(&addBalanceReq); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": err.Error(),
		})
	}
	err := service.AddBalanceByID(addBalanceReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"message": "Add balance successfully",
	})
}
func AddBalanceAllUser(c echo.Context) error {
	req := helper.Gif{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": err.Error(),
		})
	}
	//begin transaction
	tx := storage.GetDB().Begin()
	err := service.AddBalanceAllUser(tx, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"message": "Add balance for all user successfully",
	})
}

func MyProfile(c echo.Context) error {
	idInt := c.Get("userID").(uint)
	fmt.Println("id: ", idInt)
	user, err := userServices.UserDetail(uint(idInt))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	fmt.Println("user: ", user)
	return c.JSON(http.StatusOK, user)
}