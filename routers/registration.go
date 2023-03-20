package routers

import (
	"github.com/gin-gonic/gin"
	"pubsubapi/controllers"
)

func registrationRoutes(reg *gin.RouterGroup) {
	reg.POST("", controllers.RegisterUser)
	reg.GET("/unverified", controllers.ListUnverifiedUsers)
	reg.POST("/unverified", controllers.RespondUnverifiedUsers)
	reg.GET("/rejected", controllers.RejectedUsers)
}
