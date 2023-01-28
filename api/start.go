package api

import "github.com/gin-gonic/gin"

func Start() {
	router := gin.Default()
	group := router.Group("/api")
	group.POST("/pull", func(ctx *gin.Context) {
		performPulls(ctx)
	})
	router.Run(":4000")
}
