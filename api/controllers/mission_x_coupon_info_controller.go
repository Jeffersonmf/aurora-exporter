package controllers

import (
	"context"
	"net/http"
	"time"

	"brandlovrs.exporter/pkg/models"
	"github.com/gin-gonic/gin"
)

type MissionXCouponInfoController struct{}

func NewMissionXCouponInfoController() *MissionXCouponInfoController {
	return &MissionXCouponInfoController{}
}

func (bd *MissionXCouponInfoController) ImportAllRows(c *gin.Context) {

	ctx := context.Background()

	go models.MissionXCouponInfoExporter(ctx)

	c.JSON(http.StatusOK, gin.H{"Process Started at": time.Now()})
}
