package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/controllers/usecase"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/utils/codes"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

type TagController struct {
	db   db.DB
	conf app.Config
	tag  usecase.TagUsecase
}

func (tc *TagController) Init(db db.DB, conf app.Config) *TagController {
	tc.db = db
	tc.conf = conf
	tc.tag = *new(usecase.TagUsecase).Init(db, conf)

	return tc
}

func (tc *TagController) GetNodesOfTag(c *gin.Context) {
	name := strings.ToLower(c.Query("name"))
	limit := c.Query("limit")
	offset := c.Query("limit")

	tag, err := tc.tag.GetTagByName(name)

	if err != nil {
		logrus.Infof(err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": codes.TagNotFound})
		return
	}

	nodes, count, err := tc.tag.GetNodesOfTag(*tag, limit, offset)

	c.JSON(http.StatusOK, gin.H{"nodes": nodes, "count": count})
}
