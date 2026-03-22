package handlers

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
)

// SystemHandler 系统相关处理
func SystemHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"os": runtime.GOOS,
			"arch": runtime.GOARCH,
			"version": runtime.Version(),
		})
	}
}
