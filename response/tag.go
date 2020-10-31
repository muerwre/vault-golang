package response

import "github.com/muerwre/vault-golang/models"

type TagAutocomplete struct {
	Title string `json:"title"`
}

func (ta *TagAutocomplete) Init(t models.Tag) *TagAutocomplete {
	ta.Title = t.Title
	return ta
}
