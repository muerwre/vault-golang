package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/muerwre/vault-golang/internal/db"
	"github.com/muerwre/vault-golang/internal/feature/lab/request"
	"github.com/muerwre/vault-golang/internal/feature/lab/response"
	"github.com/muerwre/vault-golang/internal/feature/lab/usecase"
	"github.com/muerwre/vault-golang/pkg/codes"
)

type LabController struct {
	lab usecase.LabUsecase
}

func (ctrl *LabController) Init(db db.DB) *LabController {
	ctrl.lab = *new(usecase.LabUsecase).Init(db)
	return ctrl
}

func (ctrl LabController) List(c *gin.Context) {
	req := &request.LabListQuery{}
	_ = c.BindQuery(req)

	nodes, count, err := ctrl.lab.GetList(req.After, req.Limit)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.NodeNotFound})
		return
	}

	resp := new(response.LabListResponse).Init(nodes, count)

	c.JSON(http.StatusOK, resp)
}
