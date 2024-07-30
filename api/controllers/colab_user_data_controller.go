package controllers

import (
	"context"
	"net/http"
	"time"

	"brandlovrs.exporter/pkg/models"
	"github.com/gin-gonic/gin"
)

type ColabUserDataController struct{}

func NewColabUserDataController() *ColabUserDataController {
	return &ColabUserDataController{}
}

func (bd *ColabUserDataController) ImportAllRows(c *gin.Context) {

	ctx := context.Background()

	go models.ColabUserDataExporter(ctx)

	c.JSON(http.StatusOK, gin.H{"Process Started at": time.Now()})
}
