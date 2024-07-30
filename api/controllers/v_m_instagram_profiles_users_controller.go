package controllers

import (
	"context"
	"net/http"
	"time"

	"brandlovrs.exporter/pkg/models"
	"github.com/gin-gonic/gin"
)

type VMInstagramProfilesUsersController struct{}

func NewVMInstagramProfilesUsersController() *VMInstagramProfilesUsersController {
	return &VMInstagramProfilesUsersController{}
}

func (bd *VMInstagramProfilesUsersController) ImportAllRows(c *gin.Context) {

	ctx := context.Background()

	go models.VMInstagramProfilesUsersExporter(ctx)

	c.JSON(http.StatusOK, gin.H{"Process Started at": time.Now()})
}
