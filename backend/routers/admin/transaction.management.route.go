package adminRoute

import (
	"main/handlers/controllers/chain"

	"github.com/labstack/echo/v4"
)

func TransactionManagementRoute(router *echo.Group) {
	router.GET("/transactions", chain.GetAllTransactionHash)
	router.GET("/detail-transaction/:hash", chain.FetchTxByHash)
}
