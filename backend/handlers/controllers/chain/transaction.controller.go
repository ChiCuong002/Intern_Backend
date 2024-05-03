package chain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	service "main/handlers/services/user"
	"main/schema"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type TransactionData struct {
	SenderAddr    string `json:"sender_address"`
	RecipientAddr string `json:"recipient_address"`
	Amount        string `json:"amount"`
}

func GenerateTransaction(c echo.Context) error {
	var txData TransactionData
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
	cmd := exec.Command("docker", "exec", "checkers", "checkersd", "tx", "bank", "send", txData.SenderAddr, txData.RecipientAddr, txData.Amount, "--generate-only")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	if err := cmd.Start(); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Read the JSON output
	jsonData, err := io.ReadAll(stdout)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	if err := cmd.Wait(); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Parse the JSON output into a struct
	//var txData TransactionData
	if err := json.Unmarshal(jsonData, &txData); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Set the response headers
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
	c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename=unsigned_tx.json")

	// Send the transaction data as a JSON response
	return c.JSONBlob(http.StatusOK, jsonData)
}
func BroadcastTransaction(c echo.Context) error {
	// Get the uploaded file from the request
	file, err := c.FormFile("file")
	if err != nil {
		return c.String(http.StatusBadRequest, "No file uploaded")
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer src.Close()

	// Read the file content
	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Copy the file content to the Docker container
	if err := copyToDockerContainer(fileBytes, file.Filename); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Execute the CLI command to broadcast the signed transaction
	cmd := exec.Command("docker", "exec", "checkers", "checkersd", "tx", "broadcast", "/root/"+file.Filename, "--chain-id", "checkers", "--node", "tcp://localhost:26657")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("Error broadcasting transaction: %s\n%s", err.Error(), stderr.String()))
	}
	// Extract txhash from stdout
	lines := strings.Split(stdout.String(), "\n")
	var txHash string
	for _, line := range lines {
		if strings.HasPrefix(line, "txhash:") {
			txHash = strings.TrimSpace(strings.TrimPrefix(line, "txhash:"))
			break
		}
	}

	// Save txHash into database
	if txHash != "" {
		addTransactionHash := schema.TransactionHash{
			TransactionID:   txHash,
			UserID:          c.Get("userID").(uint),
			TransactionDate: time.Now(),
		}
		if err := service.InsertTransactionHash(addTransactionHash); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}
	// Send a success response
	return c.String(http.StatusOK, stdout.String())
}

func copyToDockerContainer(fileBytes []byte, fileName string) error {
	// Create a temporary file with the file content
	tempFile, err := os.CreateTemp("", fileName)
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write(fileBytes); err != nil {
		return err
	}

	// Copy the temporary file to the Docker container
	cmd := exec.Command("docker", "cp", tempFile.Name(), "checkers:/root/"+fileName)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
