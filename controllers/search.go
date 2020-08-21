package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/request"
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
	req := request.SearchNodeRequest{
		Text: "",
		Take: 20,
		Skip: 0,
	}

	resp := &response.SearchNodeResponse{
		Nodes: make([]response.SearchNodeResponseNode, 0),
	}

	if err := c.BindQuery(&req); err != nil {
		c.JSON(http.StatusOK, resp)
		return
	}

	req.Sanitize()

	if len(req.Text) == 0 {
		c.JSON(http.StatusOK, resp)
		return
	}

	nodes := sc.db.NodeRepository.GetForSearch(req.Text, req.Take, req.Skip)

	for _, v := range nodes {
		node := new(response.SearchNodeResponseNode).FromNode(*v)
		resp.Nodes = append(resp.Nodes, *node)
	}

	c.JSON(http.StatusOK, resp)
}
