package controllers

import (
	"context"
	"net/http"
	"time"

	"brandlovrs.exporter/pkg/models"
	"github.com/gin-gonic/gin"
)

type VMsnPostHashtagsController struct{}

func NewVMsnPostHashtagsController() *VMsnPostHashtagsController {
	return &VMsnPostHashtagsController{}
}

func (bd *VMsnPostHashtagsController) ImportAllRows(c *gin.Context) {

	ctx := context.Background()

	go models.VMsnPostHashtagsExporter(ctx)

	c.JSON(http.StatusOK, gin.H{"Process Started at": time.Now()})
}
