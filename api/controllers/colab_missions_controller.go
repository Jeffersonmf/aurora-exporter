package controllers

import (
	"context"
	"net/http"
	"time"

	"brandlovrs.exporter/pkg/models"
	"github.com/gin-gonic/gin"
)

type ColabMissionController struct{}

func NewColabMissionController() *ColabMissionController {
	return &ColabMissionController{}
}

func (bd *ColabMissionController) ImportAllRows(c *gin.Context) {

	ctx := context.Background()

	go models.ColabMissionExporter(ctx)

	c.JSON(http.StatusOK, gin.H{"Process Started at": time.Now()})
}
