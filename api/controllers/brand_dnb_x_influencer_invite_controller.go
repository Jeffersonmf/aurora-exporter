package controllers

import (
	"context"
	"net/http"
	"time"

	"brandlovrs.exporter/pkg/models"
	"github.com/gin-gonic/gin"
)

type BrandDnbXInfluencerInviteController struct{}

func NewBrandDnbXInfluencerInviteController() *BrandDnbXInfluencerInviteController {
	return &BrandDnbXInfluencerInviteController{}
}

func (bd *BrandDnbXInfluencerInviteController) ImportAllRows(c *gin.Context) {

	ctx := context.Background()

	go models.BrandDnbXInfluencerInviteExporter(ctx)

	c.JSON(http.StatusOK, gin.H{"Process Started at": time.Now()})
}
