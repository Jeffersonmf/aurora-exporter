package controllers

import (
	"context"
	"net/http"
	"time"

	"brandlovrs.exporter/pkg/models"
	"github.com/gin-gonic/gin"
)

type VMInstagramProfilesBrandsController struct{}

func NewVMInstagramProfilesBrandsController() *VMInstagramProfilesBrandsController {
	return &VMInstagramProfilesBrandsController{}
}

func (bd *VMInstagramProfilesBrandsController) ImportAllRows(c *gin.Context) {

	ctx := context.Background()

	go models.VMInstagramProfilesBrandsExporter(ctx)

	c.JSON(http.StatusOK, gin.H{"Process Started at": time.Now()})
}
