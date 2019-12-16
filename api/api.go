package api

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/models"
)

type API struct {
	App *app.App
	DB  *db.DB
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
}

func (a *API) AuthRequired(c *gin.Context) {
	re := regexp.MustCompile(`Bearer (.*)`)
	token := string(re.FindSubmatch([]byte(c.GetHeader("authorization")))[1])

	fmt.Printf("Token is %s", token)

	if token == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Empty credentials, id and token are required"})
		return
	}

	user, err := a.DB.GetUserByToken(token)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	c.Set("User", user)
	c.Next()
}

func (a *API) AuthOptional(c *gin.Context) {
	token := c.GetHeader("authorization")

	if token == "" {
		c.Set("User", &models.User{})
		c.Next()
	}

	user, err := a.DB.GetUserByToken(token)

	if err != nil {
		c.Set("User", &models.User{})
		c.Next()
	}

	c.Set("User", user)
	c.Next()
}

func (a *API) CorsHandler(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

	c.AbortWithStatus(204)
}
