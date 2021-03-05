package api

import "github.com/gin-gonic/gin"

func (a *API) CorsHandler(c *gin.Context) {
	c.AbortWithStatus(204)
}
