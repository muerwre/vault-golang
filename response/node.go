package response

import "github.com/muerwre/vault-golang/models"

type NodeRelatedResponse struct {
	Albums  map[string][]models.NodeRelatedItem `json:"albums"`
	Similar []models.NodeRelatedItem            `json:"similar"`
}
