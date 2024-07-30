package controllers

import (
	"context"
	"net/http"
	"time"

	"brandlovrs.exporter/pkg/models"
	"github.com/gin-gonic/gin"
)

type VMBrandsController struct{}

func NewVMBrandsController() *VMBrandsController {
	return &VMBrandsController{}
}

func (bd *VMBrandsController) ImportAllRows(c *gin.Context) {

	ctx := context.Background()

	go models.VMBrandsExporter(ctx)

	c.JSON(http.StatusOK, gin.H{"Process Started at": time.Now()})
}
