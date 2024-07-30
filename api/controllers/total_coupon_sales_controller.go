package controllers

import (
	"context"
	"net/http"
	"time"

	"brandlovrs.exporter/pkg/models"
	"github.com/gin-gonic/gin"
)

type TotalCouponSalesController struct{}

func NewTotalCouponSalesController() *TotalCouponSalesController {
	return &TotalCouponSalesController{}
}

func (bd *TotalCouponSalesController) ImportAllRows(c *gin.Context) {

	ctx := context.Background()

	go models.TotalCouponSalesExporter(ctx)

	c.JSON(http.StatusOK, gin.H{"Process Started at": time.Now()})
}
