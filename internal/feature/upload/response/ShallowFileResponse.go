package response

import "github.com/muerwre/vault-golang/internal/db/models"

// ShallowFileResponse is a response without useless info
type ShallowFileResponse struct {
	Id       uint                `json:"id"`
	Url      string              `json:"url"`
	Metadata models.FileMetadata `json:"metadata"`
}

func (r *ShallowFileResponse) FromModel(m *models.File) *ShallowFileResponse {
	if m == nil {
		return nil
	}

	r.Id = m.ID
	r.Url = m.Url
	r.Metadata = m.Metadata
	return r
}
