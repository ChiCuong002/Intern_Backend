package controllers

import (
	"context"
	"fmt"
	"log"
	storage "main/database"
	service "main/handlers/services/product"
	paginationHelper "main/helper/struct"
	helper "main/helper/struct/product"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type productForm struct {
	CategoryID  uint                    `json:"category_id" form:"category_id"`
	ProductName string                  `json:"product_name" form:"product_name"`
	Description string                  `json:"description" form:"description"`
	Price       float64                 `json:"price" form:"price"`
	Quantity    uint                    `json:"quantity" form:"quantity"`
	Images      []*multipart.FileHeader `json:"images" form:"images"`
}
type awsService struct {
	S3Client *s3.Client
}

func (awsSvc awsService) UploadFile(bucketName, bucketKey string, file multipart.File) error {
	// Read the first 512 bytes of the file
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil {
		log.Println("Error while reading the file ", err)
		return err
	}
	// Detect the Content-Type of the file
	contentType := http.DetectContentType(buffer)

	// Reset the read pointer to the start of the file
	file.Seek(0, 0)

	_, err = awsSvc.S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(bucketKey),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		log.Println("Error while uploading the file ", err)
	}
	return err
}
func getImagePath(bucketKey string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", os.Getenv("BUCKETNAME"), os.Getenv("REGION"), bucketKey)
}
func configureAWSService() (awsService, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("REGION")),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), "")),
	)
	if err != nil {
		return awsService{}, err
	}
	return awsService{S3Client: s3.NewFromConfig(cfg)}, nil
}
func AddProduct(c echo.Context) error {
	productData := &productForm{}
	if err := c.Bind(productData); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid request data"})
	}
	//begin transaction
	tx := storage.GetDB().Begin()
	//create product type
	product := &helper.ProductInsert{
		UserID:      c.Get("userID").(uint),
		CategoryID:  productData.CategoryID,
		ProductName: productData.ProductName,
		Description: productData.Description,
		Price:       productData.Price,
		Quantity:    productData.Quantity,
		StatusID:    2,
	}
	// //insert product into database
	err := service.InsertProduct(tx, product)
	if err != nil {
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to insert product"})
	}
	// //Get multipart form from the request
	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid form data"})
	}
	// Get the image files from the form
	images, ok := form.File["images"]
	if !ok {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "No image files"})
	}
	//load .env variable
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// Configure the AWS service
	awsService, err := configureAWSService()
	if err != nil {
		log.Println("Error while configuring the AWS service:", err)
	}
	// Iterate over the image files and upload each one
	for _, image := range images {
		// Generate a unique key for the S3 bucket
		bucketKey := uuid.New().String()
		//create image path
		imagePath := getImagePath(bucketKey)
		//insert the image into database
		img := &helper.ImageInsert{
			BucketKey: bucketKey,
			Path:      imagePath,
		}
		err = service.InsertImage(tx, img)
		if err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to insert image"})
		}
		productImg := &helper.ProductImageInsert{
			ProductID: product.ProductID,
			ImageID:   img.ImageID,
		}
		err = service.InsertProductImage(tx, productImg)
		if err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to insert product image"})
		}
		//product.Images = append(product.Images, *img)
		// Open the image file
		src, err := image.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
		defer src.Close()
		// Upload the image to AWS
		err = awsService.UploadFile(os.Getenv("BUCKETNAME"), bucketKey, src)
		if err != nil {
			log.Println("Failed to upload the image:", err)
		} else {
			log.Println("Image uploaded successfully")
		}
	}
	tx.Commit()
	return c.JSON(http.StatusOK, echo.Map{
		"message": "Add product successfully",
		"product": product,
	})
}

func DetailProduct(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "Failed to get product id",
		})
	}
	product, err := service.DetailProduct(uint(productID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"message": "success",
		"product": product,
	})
}

func UpdateProduct(c echo.Context) error {
	userID := c.Get("userID").(uint)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": err.Error(),
		})
	}
	//compare owner id
	if err := service.CompareUserID(userID, uint(id)); err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"message": err.Error(),
		})
	}
	product := helper.UpdateProduct{}
	if err := c.Bind(&product); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": err.Error(),
		})
	}
	product.ProductID = uint(id)
	err = service.UpdateProduct(&product)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"message": "Update product data successfully",
	})
}

const (
	LIMIT_DEFAULT = 10
	PAGE_DEFAULT  = 1
	SORT_DEFAULT  = " product_id desc"
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
func GetAllProduct(c echo.Context) error {
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
	pagination := paginationHelper.Pagination{
		Page:   page,
		Limit:  limit,
		Sort:   sort,
		Search: search,
	}
	products, err := service.GetAllProduct(pagination)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, products)
}
func MyInventory(c echo.Context) error {
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
	pagination := paginationHelper.Pagination{
		Page:   page,
		Limit:  limit,
		Sort:   sort,
		Search: search,
	}
	userID := c.Get("userID").(uint)
	products, err := service.GetMyInventory(pagination, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, products)
}
func MyProduct(c echo.Context) error {
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
	pagination := paginationHelper.Pagination{
		Page:   page,
		Limit:  limit,
		Sort:   sort,
		Search: search,
	}
	userID := c.Get("userID").(uint)
	products, err := service.GetMyInventory(pagination, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, products)
}
func DeActiveProduct(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "Failed to get product id. " + err.Error(),
		})
	}
	err = service.DeActiveProduct(uint(productID), c.Get("userID").(uint))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"message": "Block product successfully",
	})
}
func ActiveProduct(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "Failed to get product id",
		})
	}
	err = service.ActiveProduct(uint(productID), c.Get("userID").(uint))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"message": "Unblock product successfully",
	})
}

// order
type ThongTinOrder struct {
	//UserID    uint `json:"userID"`
	ProductID uint `json:"productID"`
}

func PurchaseProductController(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var thongtin ThongTinOrder

		// Bind JSON data from the request to thongtin
		if err := c.Bind(&thongtin); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request payload"})
		}

		// Call the original PurchaseProduct function
		err := service.PurchaseProduct(db, c, c.Get("userID").(uint), thongtin.ProductID)
		if err != nil {
			// Handle error if needed
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}

		// Return success response
		return c.JSON(http.StatusOK, echo.Map{"message": "Purchase successful"})
	}
}
func DeleteProduct(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "Failed to get product id",
		})
	}
	err = service.DeleteProduct(uint(productID), c.Get("userID").(uint))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"message": "Xóa bài đăng thành công",
	})
}
