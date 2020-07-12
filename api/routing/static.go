package routing

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/utils"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

type StaticRouter struct {
	api    utils.AppApi
	config app.Config
}

func (sr *StaticRouter) Init(a utils.AppApi, config app.Config) {
	sr.api = a
	sr.config = config
}

func (sr *StaticRouter) Handle(r *gin.RouterGroup) {
	r.Use(sr.FallbackMiddleware)
	r.Static("/", sr.config.UploadPath)
}

func (sr *StaticRouter) FallbackMiddleware(c *gin.Context) {
	re := regexp.MustCompile(`^/static/cache/([^/]+)/(.*)`)
	matches := re.FindSubmatch([]byte(c.Request.RequestURI))

	if len(matches) < 3 {
		c.Next()
		return
	}

	preset := string(matches[1])
	src := string(matches[2])
	dest := filepath.Join("cache", preset, src)

	if _, err := os.Stat(filepath.Join(sr.config.UploadPath, dest)); err == nil {
		c.Next()
		return
	}

	buff, err := utils.CreateScaledImage(
		filepath.Join(sr.config.UploadPath, src),
		filepath.Join(sr.config.UploadPath, dest),
		preset,
	)

	if err != nil {
		c.Next()
		return
	}

	c.String(http.StatusOK, buff.String())
	c.Abort()
}
