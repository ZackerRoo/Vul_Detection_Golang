package web

import (
	"backend/client"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type DetectRequest struct {
	ContractSourcecode string `json:"contractSourcecode"`
}

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	r.POST("/detect", func(c *gin.Context) {
		var req DetectRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		contractSourcecode := req.ContractSourcecode
		contractSourcecode = strings.TrimSpace(contractSourcecode)
		// log.Printf("Contract source code: %v", contractSourcecode)
		message, err := client.DetectContract(contractSourcecode)
		if err != nil {
			log.Printf("Error detecting vulnerability: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to detect vulnerability"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"result": message})
	})

	return r
}

func GinMain() {
	router := SetupRouter()
	if err := router.Run(":7070"); err != nil {
		fmt.Printf("Failed to run server: %v\n", err)
	}
}
