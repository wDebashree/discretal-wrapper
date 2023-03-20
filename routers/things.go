package routers

import (
	"github.com/gin-gonic/gin"
	"pubsubapi/controllers"
)

func thingsRoutes(things *gin.RouterGroup) {
	things.POST("", controllers.CreateThing)
	things.GET("", controllers.GetThings)
	things.GET("/:id", controllers.GetThing)
	things.PUT("/:id", controllers.UpdateThing)
	things.DELETE("/:id", controllers.DeleteThing)
	things.GET("/:id/channels", controllers.GetConnectedChannels)
	things.GET("/:id/groups", controllers.GetConnectedGroups)
	things.POST("/:id/groups", controllers.AssignGroups)
	things.DELETE("/:id/groups", controllers.UnassignGroups)
}
