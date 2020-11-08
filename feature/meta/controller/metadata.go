package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/feature/meta/usecase"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

type MetaController struct {
	meta usecase.MetaUsecase
}

func (mc *MetaController) Init(db db.DB, config app.Config) *MetaController {
	mc.meta = *new(usecase.MetaUsecase).Init(db, config)
	return mc
}

func (mc MetaController) GetYoutubeTitles(c *gin.Context) {
	ids := strings.Split(c.Query("ids"), ",")

	embeds, err := mc.meta.GetEmbedsFromDbOrFetchFromGoogle(ids)
	if err != nil {
		logrus.Warnf("Can't find info for embeds: %+v", ids)
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": embeds})
}
