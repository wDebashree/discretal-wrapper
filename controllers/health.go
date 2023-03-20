package controllers

import (
	"fmt"
	"net/http"
	"pubsubapi/env"

	"github.com/gin-gonic/gin"
)

// // GetHealth	godoc

// // 	@Summary		Health check
// // 	@Description	Retrieves service health check info.
// // 	@Tags			app health
// // 	@Produce		json
// // 	@Success		200	"Service Health Check."
// // 	@Failure		500	"Unexpected server-side error occurred."
// // 	@Router			/health [get]
func GetHealth(c *gin.Context) {
	thingsurl := env.Env(envThingURL, sdkThingURL)
	thingsurl = thingsurl + "/health"
	fmt.Println("thingsurl: ", thingsurl)

	// This will fail because requiring "authorization key"
	resp, err := httpReq(c, "GET", thingsurl, nil)
	if err != nil {
		fmt.Println("err =", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	fmt.Println("1: status = ", resp.StatusCode)

	usersurl := env.Env(envUserURL, sdkUserURL)
	usersurl = usersurl + "/health"
	fmt.Println("usersurl: ", usersurl)

	// This will fail because requiring "authorization key"
	resp, err = httpReq(c, "GET", usersurl, nil)
	if err != nil || resp.StatusCode != 200 {
		fmt.Println("2: status = ", resp.StatusCode)
		c.Status(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	c.Status(http.StatusOK)
}
