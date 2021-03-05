package routing

import (
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/internal/app"
	"github.com/muerwre/vault-golang/internal/service/image"
	"github.com/muerwre/vault-golang/pkg"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

type StaticRouter struct {
	api    pkg.AppApi
	config app.Config
}

func (sr *StaticRouter) Init(a pkg.AppApi, config app.Config) {
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

	preset, _ := url.QueryUnescape(string(matches[1]))
	src, _ := url.QueryUnescape(string(matches[2]))
	dest := filepath.Join("cache", preset, src)

	if _, err := os.Stat(filepath.Join(sr.config.UploadPath, dest)); err == nil {
		cacheSince := time.Now().Format(http.TimeFormat)
		cacheUntil := time.Now().AddDate(0, 6, 0).Format(http.TimeFormat)
		cacheMaxAge := time.Since(time.Now().AddDate(0, 6, 0)).Seconds()

		c.Header("Cache-Control", fmt.Sprintf("max-age: %d, public", -int(cacheMaxAge)))
		c.Header("Last-Modified", cacheSince)
		c.Header("Expires", cacheUntil)

		if sr.config.UploadOutputWebp {
			mime, err := mimetype.DetectFile(filepath.Join(sr.config.UploadPath, dest))

			if err != nil {
				c.Next()
				return
			}

			if mime.String() == "image/webp" {
				c.Header("Content-type", "image/webp")
			}
		}
		c.Next()
		return
	}

	mime, err := mimetype.DetectFile(filepath.Join(sr.config.UploadPath, src))
	if err != nil {
		c.Next()
		return
	}

	buff, err := image.CreateScaledImage(
		filepath.Join(sr.config.UploadPath, src),
		filepath.Join(sr.config.UploadPath, dest),
		preset,
		sr.config.UploadOutputWebp,
	)

	if err != nil {
		c.Next()
		return
	}

	c.Header("Content-Type", mime.String())
	c.String(http.StatusOK, buff.String())
	c.Abort()
}
