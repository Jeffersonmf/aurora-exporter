package controllers

import (
	"context"
	"net/http"
	"time"

	"brandlovrs.exporter/pkg/models"
	"github.com/gin-gonic/gin"
)

type VMPostMetricsController struct{}

func NewVMPostMetricsController() *VMPostMetricsController {
	return &VMPostMetricsController{}
}

func (bd *VMPostMetricsController) ImportAllRows(c *gin.Context) {

	ctx := context.Background()

	go models.VMPostMetricsExporter(ctx)

	c.JSON(http.StatusOK, gin.H{"Process Started at": time.Now()})
}
