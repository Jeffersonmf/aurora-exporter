package controllers

import (
	"context"
	"net/http"
	"time"

	"brandlovrs.exporter/pkg/models"
	"github.com/gin-gonic/gin"
)

type WorkspaceController struct{}

func NewWorkspaceController() *WorkspaceController {
	return &WorkspaceController{}
}

func (bd *WorkspaceController) ImportAllRows(c *gin.Context) {

	ctx := context.Background()

	go models.WorkspaceExporter(ctx)

	c.JSON(http.StatusOK, gin.H{"Workspace Exporter Process Started at": time.Now()})
}
