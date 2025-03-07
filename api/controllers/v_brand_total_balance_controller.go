package controllers

import (
	"context"
	"net/http"
	"time"

	"brandlovrs.exporter/pkg/models"
	"github.com/gin-gonic/gin"
)

type VBrandTotalBalanceController struct{}

func NewVBrandTotalBalanceController() *VBrandTotalBalanceController {
	return &VBrandTotalBalanceController{}
}

func (bd *VBrandTotalBalanceController) ImportAllRows(c *gin.Context) {

	ctx := context.Background()

	go models.VBrandTotalBalanceExporter(ctx)

	c.JSON(http.StatusOK, gin.H{"Process Started at": time.Now()})
}
