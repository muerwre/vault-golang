package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/db"
	response2 "github.com/muerwre/vault-golang/feature/stats/response"
	"net/http"
	"time"
)

type StatsController struct {
	DB db.DB
}

func (sc *StatsController) GetStats(c *gin.Context) {
	boris, _ := sc.DB.Node.GetNodeBoris()
	flowLastPost, _ := sc.DB.Node.GetFlowLastPost()
	images := sc.DB.Node.GetImagesCount()
	audios := sc.DB.Node.GetAudiosCount()
	videos := sc.DB.Node.GetVideosCount()
	texts := sc.DB.Node.GetTextsCount()

	stats := response2.StatsResponse{
		StatsUsers: response2.StatsUsers{
			Total: sc.DB.User.GetTotalCount(),
			Alive: sc.DB.User.GetAliveCount(),
		},
		StatsNodes: response2.StatsNodes{
			Images: images,
			Audios: audios,
			Videos: videos,
			Texts:  texts,
			Total:  images + audios + videos + texts,
		},
		StatsComments: response2.StatsComments{
			Total: sc.DB.Node.GetCommentsCount(),
		},
		StatsFiles: response2.StatsFiles{
			Count: sc.DB.File.GetTotalCount(),
			Size:  sc.DB.File.GetTotalSize(),
		},
		StatsTimestamps: response2.StatsTimestamps{
			BorisLastComment: boris.CommentedAt.Format(time.RFC3339),
			FlowLastPost:     flowLastPost.CreatedAt.Format(time.RFC3339),
		},
	}

	c.JSON(http.StatusOK, stats)
}
