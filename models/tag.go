package models

type TagData struct {
	*SimpleJson
}

type Tag struct {
	ID    uint
	Title string  `json:"title"`
	Nodes []*Node `gorm:"many2many:node_tags_tag;jointable_foreignkey:tagId;association_jointable_foreignkey:nodeId;" json:"-"`
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
