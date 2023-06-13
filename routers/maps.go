package routers

import (
	"github.com/gin-gonic/gin"
	"pubsubapi/controllers"
)

func mapsRoutes(maps *gin.RouterGroup) {
	maps.GET("", controllers.GetMaps)
	maps.GET("/:id", controllers.GetMap)
	maps.POST("", controllers.AddMaps)
	maps.PUT("/:id", controllers.UpdateMap)
	maps.DELETE("/:id", controllers.RemoveMap)
}
