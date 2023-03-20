package routers

import (
	"github.com/gin-gonic/gin"
	"pubsubapi/controllers"
)

func statusRoutes(status *gin.RouterGroup) {
	status.GET("", controllers.GetHealth)
}
