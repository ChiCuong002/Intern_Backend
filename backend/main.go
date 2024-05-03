package main

import (
	"main/configs"
	storage "main/database"
	"main/routers"

	"github.com/labstack/echo/v4"
	gomiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
	app := echo.New()
	storage.InitDB()
	storage.Migration()
	app.Use(gomiddleware.Logger())
	app.Use(gomiddleware.CORSWithConfig(configs.CORSConfig()))
	routers.Routers(app)
	app.Logger.Fatal(app.Start(":8080"))
}
