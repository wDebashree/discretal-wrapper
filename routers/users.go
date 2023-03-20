package routers

import (
	"github.com/gin-gonic/gin"
	"pubsubapi/controllers"
)

func usersRoutes(users *gin.RouterGroup) {
	users.GET("", controllers.GetUsers)
	users.PUT("", controllers.UpdateUser)
	users.GET("/profile", controllers.GetUserProfile)
	users.GET("/:id/groups", controllers.GetConnectedGroups)
	users.POST("/:id/groups", controllers.AssignGroups)
	users.DELETE("/:id/groups", controllers.UnassignGroups)
}
