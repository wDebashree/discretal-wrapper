package routers

import (
	"github.com/gin-gonic/gin"
	"pubsubapi/controllers"
)

func messagesRoutes(messages *gin.RouterGroup) {
	messages.POST("", controllers.PublishMessage)
}
