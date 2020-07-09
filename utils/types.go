package utils

import "github.com/gin-gonic/gin"

type AppApi interface {
	AuthRequired(c *gin.Context)
	AuthOptional(c *gin.Context)
	WithUser(preload bool) func(*gin.Context)
}
