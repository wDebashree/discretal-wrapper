package logger

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

func LoggingMiddleware(params gin.LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if params.IsOutputColor() {
		statusColor = params.StatusCodeColor()
		methodColor = params.MethodColor()
		resetColor = params.ResetColor()
	}

	if params.Latency > time.Minute {
		params.Latency = params.Latency.Truncate(time.Second)
	}
	// return fmt.Sprintf("Hi %v |%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s\n%s",
	// 	params.TimeStamp.Format("2006-01-02T15:04:05.000000000z"),
	// 	statusColor, params.StatusCode, resetColor,
	// 	params.Latency,
	// 	params.ClientIP,
	// 	methodColor, params.Method, resetColor,
	// 	params.Path,
	// 	params.ErrorMessage,
	// 	params.Keys["level"],
	// )

	if _, ok := params.Keys["level"]; !ok {
		params.Keys = make(map[string]any)
		params.Keys["level"] = ""
	}
	return fmt.Sprintf("{\"level\":\"%s\",\"message\":\"%s\",\"ts\":\"%v\" |%s %d %s| %v | %s |%s %s %s %#v}\n",
		params.Keys["level"],
		strings.TrimSpace(strings.TrimPrefix(params.ErrorMessage, "Error #01: ")),
		params.TimeStamp.Format("2006-01-02 15:04:05"),
		statusColor, params.StatusCode, resetColor,
		params.Latency,
		params.ClientIP,
		methodColor, params.Method, resetColor,
		params.Path,
	)
}
