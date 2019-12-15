package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/app"
)

type API struct {
	App *app.App

	// Config *Config
}

type ErrorCode struct {
	Code   string   `json:"code"`
	Stack  []string `json:"stack"`
	Reason string   `json:"reason"`
}

type statusCodeRecorder struct {
	http.ResponseWriter
	http.Hijacker
	StatusCode int
}

func New(a *app.App) (api *API, err error) {
	api = &API{App: a}

	// api.Config, err = InitConfig()

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
		// c.Set("DB", a.App.DB)
		c.Set("Config", a.App.Config)
		c.Next()
	})

	r.OPTIONS("/*path", a.CorsHandler)

	// AuthRouter(r.Group("/auth"), a)
	// RouteRouter(r.Group("/route"), a)
}

func (a *API) AuthRequired(c *gin.Context) {
	token := c.GetHeader("authorization")

	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Empty credentials, id and token are required"})
		return
	}

	// user, err := a.App.DB.GetUserByToken(token)

	// if err != nil {
	// c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
	// return
	// }

	// c.Set("User", user)
	c.Next()
}

func (a *API) AuthOptional(c *gin.Context) {
	token := c.GetHeader("authorization")

	if token == "" {
		// c.Set("User", &model.User{})
		c.Next()
	}

	// user, err := a.App.DB.GetUserByToken(token)

	// if err != nil {
		// c.Set("User", &model.User{})
		// c.Next()
	// }

	// c.Set("User", user)
	c.Next()
}

func (a *API) CorsHandler(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

	c.AbortWithStatus(204)
}
