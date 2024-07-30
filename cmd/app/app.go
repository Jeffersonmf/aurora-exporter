package app

import (
	"context"
	"runtime"

	"brandlovrs.exporter/api"
	"brandlovrs.exporter/pkg/models"
)

func InitializeRuntimers(args []string) {
	runtime.GOMAXPROCS(2)
	//go api.LauncherGin()

	ctx := context.Background()
	models.InitializeWorkers(ctx)

	go api.StartRouters(ctx)

	select {}
}
