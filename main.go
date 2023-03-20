package main

import (
	"github.com/gin-gonic/gin"
	_ "pubsubapi/docs"
	"pubsubapi/logger"
	"pubsubapi/routers"
)

//	@title			Discretal API
//	@version		1.0
//	@description	A wrapper api for utilizing Discretal server messaging services over MQTT
//	@termsOfService	http://iot.discretal.com/terms/

// @host						localhost:5000
// @BasePath					/api
// @schemes					http
//
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
func main() {
	router := gin.New()
	router.Use(gin.LoggerWithFormatter(logger.LoggingMiddleware))
	routers.BuildRoutes(router)
	router.Run("0.0.0.0:5001")
}
