package routers

import (
	"github.com/gin-gonic/gin"
	"pubsubapi/controllers"
)

func channelsRoutes(channels *gin.RouterGroup) {
	channels.POST("", controllers.CreateChannel)
	channels.GET("", controllers.GetChannels)
	channels.GET("/:id", controllers.GetChannel)
	channels.DELETE("/:id", controllers.DeleteChannel)
	channels.GET("/:id/things", controllers.GetConnectedThings)
	channels.GET("/:id/messages", controllers.GetMessages)
	channels.POST("/:id/messages", controllers.SendMessages)
	channels.POST("/:id/messages/*subtopics", controllers.SendMessages)
	// channels.GET("/:id/data", controllers.GetMapData) // Not in use unless maps updated over channels
}
