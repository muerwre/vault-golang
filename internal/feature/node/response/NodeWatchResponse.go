package response

import "github.com/muerwre/vault-golang/internal/db/models"

const (
	NodeWatchStatusActive   string = "active"
	NodeWatchStatusDisabled string = "disabled"
	NodeWatchStatusUnset    string = "unset"
)

type NodeWatchResponse struct {
	ID     uint   `json:"id"`
	Status string `json:"status"`
}

func (r *NodeWatchResponse) Init(w *models.NodeWatch) *NodeWatchResponse {
	r.ID = w.ID

	switch {
	case w == nil || w.ID == 0:
		r.Status = NodeWatchStatusUnset
	case w.Active:
		r.Status = NodeWatchStatusActive
	default:
		r.Status = NodeWatchStatusDisabled
	}

	return r
}
