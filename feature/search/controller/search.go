package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/db"
	request2 "github.com/muerwre/vault-golang/feature/search/request"
	response2 "github.com/muerwre/vault-golang/feature/search/response"
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
	req := request2.SearchNodeRequest{
		Text: "",
		Take: 20,
		Skip: 0,
	}

	resp := &response2.SearchNodeResponse{
		Nodes: make([]response2.SearchNodeResponseNode, 0),
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

	nodes, count := sc.db.Node.GetForSearch(req.Text, req.Take, req.Skip)

	for _, v := range nodes {
		node := new(response2.SearchNodeResponseNode).Init(*v)
		resp.Nodes = append(resp.Nodes, *node)
	}

	resp.Total = count

	c.JSON(http.StatusOK, resp)
}
