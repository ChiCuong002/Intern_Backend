package chain

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"main/handlers/services/transaction"
	service "main/handlers/services/user"
	helper "main/helper/struct"
	paginationHelper "main/helper/struct"
	"main/schema"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	ADMIN_ADDRESS = "cosmos1gntezce9llrjc0tqja6ehsjpt2nac4vuzy6mda"
	LIMIT_DEFAULT = 10
	PAGE_DEFAULT  = 1
	SORT_DEFAULT  = " transaction_date desc"
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

type AddressData struct {
	AccountAddress string `json:"address"`
}
type txHashData struct {
	TxHash string `json:"tx_hash"`
}

func FetchTxByHash(c echo.Context) error {
	// Replace with your actual node's URL
	nodeURL := "http://0.0.0.0:1317"
	var req txHashData
	req.TxHash = c.Param("hash")
	// if err := c.Bind(&req); err != nil {
	// 	return c.JSON(http.StatusInternalServerError, echo.Map{
	// 		"message": err.Error(),
	// 	})
	// }

	queryEndpoint := fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", nodeURL, req.TxHash)
	fmt.Println("endpoint: ", queryEndpoint)
	resp, err := http.Get(queryEndpoint)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}

	return c.JSONBlob(http.StatusOK, body)
}

func AccountBalance(c echo.Context) error {
	// Replace with your actual node's URL
	nodeURL := "http://0.0.0.0:1317"

	var req AddressData
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}
	fmt.Println("req: ", req.AccountAddress)
	// Replace with the account address you want to query
	accountAddress := req.AccountAddress

	fmt.Println("address: ", accountAddress)

	// Construct the API endpoint for querying balances
	balanceEndpoint := fmt.Sprintf("%s/bank/balances/%s", nodeURL, accountAddress)

	// Make an HTTP GET request to the endpoint
	resp, err := http.Get(balanceEndpoint)
	if err != nil {
		fmt.Println("Error:", err)
		return fmt.Errorf("")
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return fmt.Errorf("")
	}

	// Print the balance response (assuming it's in JSON format)
	fmt.Println("Account balance response:")
	fmt.Println(string(body))
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		// Handle error
		fmt.Println("Error:", err)
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{
		"data": result,
	})
}

type TransactionRequest struct {
	SenderAddr    string `json:"sender_address"`
	RecipientAddr string `json:"recipient_address"`
	Amount        string `json:"amount"`
}

func SendTx(c echo.Context) error {
	var txData TransactionRequest
	if err := c.Bind(&txData); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}
	if txData.RecipientAddr != ADMIN_ADDRESS {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "You must send money to admin address to recieve tokens",
		})
	}
	cmd := exec.Command("docker", "exec", "checkers", "checkersd", "tx", "bank", "send", txData.SenderAddr, txData.RecipientAddr, txData.Amount, "--log_format", "json", "-y")
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println("stdout Error:", err)
		if exitError, ok := err.(*exec.ExitError); ok {
			fmt.Println("Error output:", string(exitError.Stderr))
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
		})
	}

	// Split the output into lines
	lines := strings.Split(string(stdout), "\n")

	// Extract raw_log and txhash from the lines
	var rawLog, txHash string
	for _, line := range lines {
		if strings.HasPrefix(line, "raw_log:") {
			rawLog = strings.TrimSpace(strings.TrimPrefix(line, "raw_log:"))
			// Remove the outer quotes and escape characters
			rawLog = strings.TrimPrefix(rawLog, "'")
			rawLog = strings.TrimSuffix(rawLog, "'")
			rawLog = strings.ReplaceAll(rawLog, "\\\"", "\"")
		}
		if strings.HasPrefix(line, "txhash:") {
			txHash = strings.TrimSpace(strings.TrimPrefix(line, "txhash:"))
		}
	}

	//update user balance
	// Remove the "stake" part
	numericPart := strings.TrimSuffix(txData.Amount, "stake")

	// Convert the numeric part to float64
	value, err := strconv.ParseFloat(numericPart, 64)
	if err != nil {
		fmt.Println("Error:", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
			"raw_log": rawLog,
			"txhash":  txHash,
		})
	}
	addbalanceStruct := helper.AddBalanceReq{}
	addbalanceStruct.UserID = c.Get("userID").(uint)
	addbalanceStruct.AmountToAdd = value
	if err := service.AddBalanceByID(addbalanceStruct); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
			"raw_log": rawLog,
			"txhash":  txHash,
		})
	}
	//insert transaction hash
	addTransactionHash := schema.TransactionHash{}
	addTransactionHash.TransactionID = txHash
	addTransactionHash.UserID = c.Get("userID").(uint)
	addTransactionHash.TransactionDate = time.Now()
	if err := service.InsertTransactionHash(addTransactionHash); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": err.Error(),
			"raw_log": rawLog,
			"txhash":  txHash,
		})
	}
	// Return the extracted fields in the response
	return c.JSON(http.StatusOK, echo.Map{
		"message": "Send transaction successfully",
		"raw_log": rawLog,
		"txhash":  txHash,
	})
}
func GetAllTransactionHash(c echo.Context) error {
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
	transactions, err := transaction.GetAllTransactionHash(pagination)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, transactions)
}
func GetUserTransactionHash(c echo.Context) error {
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
	transactions, err := transaction.GetUserTransactionHash(pagination, c.Get("userID").(uint))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, transactions)
}
