package controllers

import (
	"context"
	"net/http"
	"time"

	"brandlovrs.exporter/pkg/models"
	"github.com/gin-gonic/gin"
)

type ColabLevelsXColabMissionsXColabRewardController struct{}

func NewColabLevelsXColabMissionsXColabRewardController() *ColabLevelsXColabMissionsXColabRewardController {
	return &ColabLevelsXColabMissionsXColabRewardController{}
}

func (bd *ColabLevelsXColabMissionsXColabRewardController) ImportAllRows(c *gin.Context) {

	ctx := context.Background()

	go models.ColabLevelsXColabMissionsXColabRewardExporter(ctx)

	c.JSON(http.StatusOK, gin.H{"Process Started at": time.Now()})
}
