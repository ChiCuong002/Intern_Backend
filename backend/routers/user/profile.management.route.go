package userRoute

import (
	"main/handlers/controllers/chain"
	chainControllers "main/handlers/controllers/chain"
	controllers "main/handlers/controllers/user"

	"github.com/labstack/echo/v4"
)

func ProfileRouters(router *echo.Group) {
	//get user profile
	router.GET("/my-profile", controllers.MyProfile)
	//update user profile
	router.PATCH("/update-profile", controllers.UpdateUser)
	//order history
	router.GET("/order-history", controllers.BlockUser)
	//block chain send transaction
	router.POST("/send-transaction", chainControllers.SendTx)
	//generate transaction
	router.POST("/generate-transaction", chainControllers.GenerateTransaction)
	//broadcast transaction
	router.POST("/broadcast-transaction", chainControllers.BroadcastTransaction)
	//user's transaction
	router.GET("/my-transactions", chainControllers.GetUserTransactionHash)
	//
	router.GET("/detail-transaction/:hash", chain.FetchTxByHash)
}
