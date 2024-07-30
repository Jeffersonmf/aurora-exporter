package controllers

import (
	"context"
	"net/http"
	"time"

	"brandlovrs.exporter/pkg/models"
	"github.com/gin-gonic/gin"
)

type VMBrandWalletHistoryController struct{}

func NewVMBrandWalletHistoryController() *VMBrandWalletHistoryController {
	return &VMBrandWalletHistoryController{}
}

func (bd *VMBrandWalletHistoryController) ImportAllRows(c *gin.Context) {

	ctx := context.Background()

	go models.VMBrandWalletHistoryExporter(ctx)

	c.JSON(http.StatusOK, gin.H{"Process Started at": time.Now()})
}
