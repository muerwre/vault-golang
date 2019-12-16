package models

type TagData struct {
	*SimpleJson
}

type Tag struct {
	ID uint

	Title string  `json:"title"`
	Data  TagData `gorm:"type:longtext" json:"-"`

	User   User `json:"-" gorm:"foreignkey:UserID"`
	UserID uint `gorm:"column:userId" json:"-"`

	Nodes []*Node `gorm:"many2many:node_tags_tag;association_jointable_foreignkey:tagId;" json:"-"`
}

func (Tag) TableName() string {
	return "tag"
}
