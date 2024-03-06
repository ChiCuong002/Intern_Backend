package normalRoute

import (
	controllers "main/handlers/controllers/normal"
	productControllers "main/handlers/controllers/product"

	"github.com/labstack/echo/v4"
)

func InitNormalRouters(app *echo.Echo) {
	//register route
	app.POST("/register", controllers.RegisterUser)
	//login route
	app.POST("/login", controllers.Login)
	//categories drop down
	app.GET("/categories-dropdown", controllers.CategoriesDropDown)
	//home page
	app.GET("/homepage", controllers.HomePage)
	//detail product
	app.GET("/product/:id", productControllers.DetailProduct)
}
