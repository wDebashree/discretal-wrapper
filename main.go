package main

import (
	_ "pubsubapi/docs"
	"pubsubapi/logger"
	"pubsubapi/routers"

	"github.com/gin-gonic/gin"
)

//	@title			Discretal API
//	@version		1.0
//	@description	A wrapper api for utilizing Discretal server messaging services over MQTT
//	@termsOfService	http://iot.discretal.com/terms/

// @host						iot.discretal.com
// @BasePath					/api
// @schemes					https
//
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
func main() {
	router := gin.New()
	router.Use(gin.LoggerWithFormatter(logger.LoggingMiddleware))
	routers.BuildRoutes(router)
	router.Run("0.0.0.0:5000")
	// if err := router.RunTLS(":5001", "certs/server.crt", "certs/server.key"); err != nil {
	// 	fmt.Println(err)
	// }
}
