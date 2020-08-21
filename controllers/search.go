package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/response"
	"net/http"
)

type SearchController struct {
	db db.DB
}

func (sc *SearchController) Init(db db.DB) *SearchController {
	sc.db = db
	return sc
}

func (sc SearchController) SearchNodes(c *gin.Context) {
	resp := &response.SearchNodeResponse{
		Total: 0,
		Nodes: []response.SearchNodeResponseNode{},
	}

	c.JSON(http.StatusOK, resp)
}
