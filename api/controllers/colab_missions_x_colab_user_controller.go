package controllers

import (
	"context"
	"net/http"
	"time"

	"brandlovrs.exporter/pkg/models"
	"github.com/gin-gonic/gin"
)

type ColabMissionsXColabUserController struct{}

func NewColabMissionsXColabUserController() *ColabMissionsXColabUserController {
	return &ColabMissionsXColabUserController{}
}

func (bd *ColabMissionsXColabUserController) ImportAllRows(c *gin.Context) {

	ctx := context.Background()

	go models.ColabMissionsXColabUserExporter(ctx)

	c.JSON(http.StatusOK, gin.H{"Process Started at": time.Now()})
}
