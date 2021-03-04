package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/internal/db"
	request2 "github.com/muerwre/vault-golang/internal/feature/search/request"
	response2 "github.com/muerwre/vault-golang/internal/feature/search/response"
	"github.com/muerwre/vault-golang/internal/feature/search/usecase"
	"net/http"
)

type SearchController struct {
	search usecase.SearchUsecase
}

func (sc *SearchController) Init(db db.DB) *SearchController {
	sc.search = *new(usecase.SearchUsecase).Init(db)
	return sc
}

func (sc SearchController) SearchNodes(c *gin.Context) {
	req := &request2.SearchNodeRequest{}
	if err := c.BindQuery(&req); err != nil {
		c.JSON(http.StatusOK, make([]response2.SearchNodeResponseNode, 0))
		return
	}
	req.Sanitize()

	resp := sc.search.GetNodesForSearch(*req)

	c.JSON(http.StatusOK, resp)
}
