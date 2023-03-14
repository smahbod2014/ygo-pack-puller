package api

import (
	"os"
	"smahbod2014/ygo-pack-puller/ui"

	"github.com/gin-gonic/gin"
)

func Start() {
	router := gin.Default()

	group := router.Group("/api")

	group.POST("/pull", func(ctx *gin.Context) {
		PerformPullsHandler(ctx)
	})

	group.GET("/packs", func(ctx *gin.Context) {
		GetPacksHandler(ctx)
	})

	ui.AddRoutes(router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}
	router.Run(":" + port)
}
