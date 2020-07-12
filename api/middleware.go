package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
	"net/http"
)

func (a *API) WithUser(preload bool) func(*gin.Context) {
	return func(c *gin.Context) {
		uid := c.MustGet("UID").(*uint)
		d := a.db

		if *uid == 0 {
			c.Set("User", nil)
		}

		user := &models.User{}
		q := d.Model(&user)

		if preload {
			q = q.Preload("Photo").Preload("Cover")
		}

		q.First(&user, "id = ?", uid)

		if user == nil || user.ID == 0 {
			c.Set("User", &models.User{ID: 0, Role: models.USER_ROLES.GUEST})
			c.Next()
		}

		c.Set("User", user)
		c.Next()
	}
}

func (a API) RecoverMiddleware(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{"details": fmt.Sprint(r), "error": codes.UnexpectedBehavior},
			)
		}
	}()

	c.Next()
}

func (a API) InjectContextMiddleware(c *gin.Context) {
	c.Set("DB", a.db)
	c.Set("Config", a.Config)
	c.Set("Mailer", a.mailer.Chan)

	c.Next()
}

func (a API) OptionsRespondMiddleware(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	c.Next()
}
