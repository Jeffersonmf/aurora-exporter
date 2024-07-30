package controllers

import (
	"context"
	"net/http"
	"time"

	"brandlovrs.exporter/pkg/models"
	"github.com/gin-gonic/gin"
)

type ColabCouponController struct{}

func NewColabCouponController() *ColabCouponController {
	return &ColabCouponController{}
}

func (bd *ColabCouponController) ImportAllRows(c *gin.Context) {

	ctx := context.Background()

	go models.ColabCouponExporter(ctx)

	c.JSON(http.StatusOK, gin.H{"Process Started at": time.Now()})
}
