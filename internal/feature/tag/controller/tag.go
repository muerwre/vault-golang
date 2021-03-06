package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/internal/app"
	"github.com/muerwre/vault-golang/internal/db"
	usecase2 "github.com/muerwre/vault-golang/internal/feature/tag/usecase"
	"github.com/muerwre/vault-golang/pkg/codes"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"strings"
)

type TagController struct {
	db   db.DB
	conf app.Config
	tag  usecase2.TagUsecase
}

func (tc *TagController) Init(db db.DB, conf app.Config) *TagController {
	tc.db = db
	tc.conf = conf
	tc.tag = *new(usecase2.TagUsecase).Init(db)

	return tc
}

func (tc TagController) GetNodesOfTag(c *gin.Context) {
	name, err := url.QueryUnescape(strings.ToLower(c.Query("name")))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": codes.TagNotFound})
		return
	}

	tag, err := tc.tag.GetTagByName(name)
	if err != nil {
		logrus.Infof(err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": codes.TagNotFound})
		return
	}

	nodes, count, err := tc.tag.GetNodesOfTag(*tag, c.Query("limit"), c.Query("offset"))

	c.JSON(http.StatusOK, gin.H{"nodes": nodes, "count": count})
}

func (tc TagController) GetAutocomplete(c *gin.Context) {
	tags, err := tc.tag.GetTagsForAutocomplete(c.Query("search"))

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"tags": []string{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tags": tags})
}
