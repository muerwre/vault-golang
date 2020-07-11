package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/response"
	"net/http"
	"time"
)

type StatsController struct {
	DB db.DB
}

func (sc *StatsController) GetStats(c *gin.Context) {
	boris, _ := sc.DB.NodeRepository.GetNodeBoris()
	flowLastPost, _ := sc.DB.NodeRepository.GetFlowLastPost()

	stats := response.StatsResponse{
		StatsUsers: response.StatsUsers{
			Total: sc.DB.UserRepository.GetTotalCount(),
			Alive: sc.DB.UserRepository.GetAliveCount(),
		},
		StatsNodes: response.StatsNodes{
			Images: sc.DB.NodeRepository.GetImagesCount(),
			Audios: sc.DB.NodeRepository.GetAudiosCount(),
			Videos: sc.DB.NodeRepository.GetVideosCount(),
			Texts:  sc.DB.NodeRepository.GetTextsCount(),
		},
		StatsComments: response.StatsComments{
			Total: sc.DB.NodeRepository.GetCommentsCount(),
		},
		StatsFiles: response.StatsFiles{
			Count: sc.DB.FileRepository.GetTotalCount(),
			Size:  sc.DB.FileRepository.GetTotalSize(),
		},
		StatsTimestamps: response.StatsTimestamps{
			BorisLastComment: boris.CommentedAt.Format(time.RFC3339),
			FlowLastPost:     flowLastPost.CreatedAt.String(),
		},
	}

	c.JSON(http.StatusOK, stats)
}
