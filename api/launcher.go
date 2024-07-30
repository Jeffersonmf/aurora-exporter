package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"brandlovrs.exporter/pkg/util"
) // gin-swagger middleware
// swagger embed files

func LauncherGin() {
	// programmatically set swagger info
	// docs.SwaggerInfo.Title = "Swagger Example API"
	// docs.SwaggerInfo.Description = "This is a sample server Petstore server."
	// docs.SwaggerInfo.Version = "1.0"
	// docs.SwaggerInfo.Host = "petstore.swagger.io"
	// docs.SwaggerInfo.BasePath = "/v2"
	// docs.SwaggerInfo.Schemes = []string{"http", "https"}

	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// use ginSwagger middleware to serve the API docs
	// router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "alive",
		})
	})

	err := router.Run(":9888") // listen and serve on 0.0.0.0:9888 (for windows "localhost:9888")
	if err != nil {
		util.Sugar.Infof("Failed to listen and serve on 0.0.0.0:9888")
	}
}
