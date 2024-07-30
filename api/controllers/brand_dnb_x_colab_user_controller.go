package controllers

import (
	"context"
	"net/http"
	"time"

	"brandlovrs.exporter/pkg/models"
	"github.com/gin-gonic/gin"
)

type BrandDnbXColabUserController struct{}

func NewBrandDnbXColabUserController() *BrandDnbXColabUserController {
	return &BrandDnbXColabUserController{}
}

func (bd *BrandDnbXColabUserController) ImportAllRows(c *gin.Context) {

	ctx := context.Background()

	go models.BrandDnbXColabUserExporter(ctx)

	c.JSON(http.StatusOK, gin.H{"Process Started at": time.Now()})
}
