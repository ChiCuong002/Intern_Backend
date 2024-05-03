package adminRoute

import (
	"main/configs"
	"main/middleware"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

func InitAdminRoute(app *echo.Echo) {
	//group route
	group := app.Group("/admin")
	//config
	group.Use(echojwt.WithConfig(configs.EchoJWTConfig()))
	group.Use(middleware.AdminAuthentication)
	//user management routers
	UserManagementRouters(group)
	//category management routers
	CategoriesManagementRouters(group)
	//product management routers
	ProductManagementRoutes(group)
	//
	TransactionManagementRoute(group)
}
