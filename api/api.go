package api

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/utils/codes"
)

type API struct {
	App    *app.App
	DB     *db.DB
	Errors map[string]string
}

type ErrorCode struct {
	Code   string   `json:"code"`
	Stack  []string `json:"stack"`
	Reason string   `json:"reason"`
}

func New(a *app.App) (api *API, err error) {
	api = &API{App: a, DB: a.DB}

	if err != nil {
		return nil, err
	}

	return api, nil
}

func (a *API) Init(r *gin.RouterGroup) {
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	r.Use(func(c *gin.Context) {
		c.Set("DB", a.App.DB)
		c.Set("Config", a.App.Config)
		c.Next()
	})

	r.OPTIONS("/*path", a.CorsHandler)

	UserRouter(r.Group("/user"), a)
	NodeRouter(r.Group("/node"), a)
	NodesRouter(r.Group("/nodes"), a)
}

func (a *API) AuthRequired(c *gin.Context) {
	re := regexp.MustCompile(`Bearer (.*)`)
	t := string(re.FindSubmatch([]byte(c.GetHeader("authorization")))[1])
	d := c.MustGet("DB").(*db.DB)

	token := &models.Token{}
	d.First(&token, "token = ?", t)

	if token.ID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": codes.EMPTY_REQUEST})
		return
	}

	c.Set("UID", token.UserID)
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
	d := c.MustGet("DB").(*db.DB)

	token := &models.Token{}
	d.First(&token, "token = ?", t)

	if token.ID == 0 {
		c.Set("UID", uint(0))
		c.Next()
		return
	}

	c.Set("UID", token.UserID)
	c.Next()
}

func (a *API) CorsHandler(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

	c.AbortWithStatus(204)
}

func (a *API) WithUser(preload bool) func(*gin.Context) {
	return func(c *gin.Context) {
		uid := c.MustGet("UID").(uint)

		if uid == 0 {
			c.Set("User", nil)
		}

		d := c.MustGet("DB").(*db.DB)

		user := &models.User{}
		q := d.Model(&user)

		if preload {
			q = q.Preload("Photo").Preload("Cover")
		}

		q.First(&user, "id = ?", uid)

		if user.ID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": codes.USER_NOT_FOUND})
			return
		}

		c.Set("User", user)
		c.Next()
	}
}
