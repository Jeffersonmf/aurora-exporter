package controllers

import (
	"context"
	"net/http"
	"time"

	"brandlovrs.exporter/pkg/models"
	"github.com/gin-gonic/gin"
)

type CalendarController struct{}

func NewCalendarController() *CalendarController {
	return &CalendarController{}
}

func (bd *CalendarController) ImportAllRows(c *gin.Context) {

	ctx := context.Background()

	go models.CalendarExporter(ctx)

	c.JSON(http.StatusOK, gin.H{"Process Started at": time.Now()})
}
