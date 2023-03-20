package routers

import (
	"github.com/gin-gonic/gin"
	"pubsubapi/controllers"
)

func loginRoutes(login *gin.RouterGroup) {
	login.POST("", controllers.LoginUser)
}
