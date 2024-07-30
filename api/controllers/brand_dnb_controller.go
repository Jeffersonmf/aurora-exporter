package controllers

import (
	"context"
	"net/http"
	"time"

	"brandlovrs.exporter/pkg/models"
	"github.com/gin-gonic/gin"
)

type BrandDnbController struct{}

func NewBrandDnbController() *BrandDnbController {
	return &BrandDnbController{}
}

func (bd *BrandDnbController) ImportAllRows(c *gin.Context) {

	ctx := context.Background()

	go models.BrandDnbExporter(ctx)

	c.JSON(http.StatusOK, gin.H{"Process Started at": time.Now()})
}
