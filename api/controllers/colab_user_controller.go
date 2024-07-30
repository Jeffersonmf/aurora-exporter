package controllers

import (
	"context"
	"net/http"
	"time"

	"brandlovrs.exporter/pkg/models"
	"github.com/gin-gonic/gin"
)

type ColabUserController struct{}

func NewColabUserController() *ColabUserController {
	return &ColabUserController{}
}

func (bd *ColabUserController) ImportAllRows(c *gin.Context) {

	ctx := context.Background()

	go models.ColabUserExporter(ctx)

	c.JSON(http.StatusOK, gin.H{"Process Started at": time.Now()})
}
