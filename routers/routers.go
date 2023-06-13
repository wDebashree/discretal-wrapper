package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
)

// var (
// 	url  = ginSwagger.URL("http://localhost:5000/swagger/doc.json")  // The url pointing to API definition
// )

func BuildRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		// router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
		router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		// api.Group("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		loginRoutes(api.Group("/login"))
		registrationRoutes(api.Group("/register"))
		thingsRoutes(api.Group("/things"))
		channelsRoutes(api.Group("/channels"))
		messagesRoutes(api.Group("/messages"))
		usersRoutes(api.Group("/users"))
		groupsRoutes(api.Group("/groups"))
		mapsRoutes(api.Group("/maps"))
		// statusRoutes(api.Group("/health"))
	}
}
