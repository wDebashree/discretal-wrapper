package errors

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrHandler(c *gin.Context, errmsg error, code int) {

	if code == http.StatusInternalServerError {
		c.JSON(code, IntServerErr{
			Error: fmt.Sprintf("%v", errmsg),
		})
	} else if code == http.StatusBadRequest {
		c.JSON(code, BadReqErr{
			Error: fmt.Sprintf("%v", errmsg),
		})
	} else {
		c.JSON(code, UnauthorizedReqErr{
			Error: fmt.Sprintf("%v", errmsg),
		})
	}
	c.Keys = make(map[string]any)
	c.Keys["level"] = "error"
	c.Error(errmsg)
}

func InfoHandler(c *gin.Context, data string, resp interface{}, code int) {
	// fmt.Printf("%+v \n", resp)
	c.JSON(code, resp)
	c.Keys = make(map[string]any)
	c.Keys["level"] = "info"
	c.Error(fmt.Errorf(data))
}

type BadReqErr struct {
	Error string `json:"error"`
}

type NotFoundErr struct {
	Error string `json:"error"`
}

type UnauthorizedReqErr struct {
	Error string `json:"error"`
}

type IntServerErr struct {
	Error string `json:"error"`
}

// type HTTPInfo struct {
// 	Result interface{} `json:"result"`
// }
