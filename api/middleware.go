package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/db/models"
	"github.com/muerwre/vault-golang/utils/codes"
	"github.com/sirupsen/logrus"
	"net/http"
	"runtime/debug"
)

func (a *API) WithUser(preloadAvatarAndCover bool) func(*gin.Context) {
	return func(c *gin.Context) {
		uid := c.MustGet("UID").(uint)
		d := a.db

		if uid == 0 {
			c.Set("User", nil)
		}

		user := &models.User{}
		q := d.Model(&user)

		if preloadAvatarAndCover {
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
			logrus.Warnf("Runtime error: %s", fmt.Sprint(r))

			println(string(debug.Stack()))

			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{"details": fmt.Sprint(r), "error": codes.UnexpectedBehavior},
			)
		}
	}()

	c.Next()
}

func (a API) InjectContextMiddleware(c *gin.Context) {
	c.Set("db", a.db)
	c.Set("config", a.Config)
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
