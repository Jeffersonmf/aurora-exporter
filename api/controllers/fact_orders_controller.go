package controllers

import (
	"context"
	"net/http"
	"time"

	"brandlovrs.exporter/pkg/models"
	"github.com/gin-gonic/gin"
)

type FactOrdersController struct{}

func NewFactOrdersController() *FactOrdersController {
	return &FactOrdersController{}
}

func (bd *FactOrdersController) ImportAllRows(c *gin.Context) {

	ctx := context.Background()

	go models.FactOrdersExporter(ctx)

	c.JSON(http.StatusOK, gin.H{"Process Started at": time.Now()})
}
