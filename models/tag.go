package models

type TagData struct {
	*SimpleJson
}

type Tag struct {
	ID    uint
	Title string `json:"title"`
}

func (Tag) TableName() string {
	return "tag"
}

func TagArrayContains(s []*Tag, el string) bool {
	for _, v := range s {
		if v.Title == el {
			return true
		}
	}

	return false
}
