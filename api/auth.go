package api

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
	"net/http"
	"regexp"
)

func (a API) AuthRequired(c *gin.Context) {
	re := regexp.MustCompile(`Bearer (.*)`)
	d := a.db

	matches := re.FindSubmatch([]byte(c.GetHeader("authorization")))

	if len(matches) < 1 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": codes.UserNotFound})
		return
	}

	t := string(matches[1])

	token := &models.Token{}
	d.First(&token, "token = ?", t)

	if token.ID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": codes.UserNotFound})
		return
	}

	c.Set("UID", *token.UserID)
	c.Next()
}

func (a *API) AuthOptional(c *gin.Context) {
	re := regexp.MustCompile(`Bearer (.*)`)
	matches := re.FindSubmatch([]byte(c.GetHeader("authorization")))

	if len(matches) < 1 {
		c.Set("UID", uint(0))
		c.Next()
		return
	}

	t := string(matches[1])
	d := a.db

	token := &models.Token{}
	d.First(&token, "token = ?", t)

	if token.ID == 0 {
		c.Set("UID", uint(0))
		c.Next()
		return
	}

	c.Set("UID", *token.UserID)
	c.Next()
}
