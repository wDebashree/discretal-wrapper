package routers

import (
	"github.com/gin-gonic/gin"
	"pubsubapi/controllers"
)

func groupsRoutes(groups *gin.RouterGroup) {
	// group related tasks
	groups.POST("", controllers.CreateGroup)
	groups.GET("", controllers.GetGroups)
	groups.GET("/:groupID", controllers.GetGroup)
	groups.DELETE("/:groupID", controllers.DeleteGroup)

	// member related tasks
	groups.POST("/:groupID/members", controllers.AddMembers)
	groups.DELETE("/:groupID/members", controllers.DeleteMembers)
	groups.GET("/:groupID/members", controllers.GetMembers)
}
