package models

type TagData struct {
	*SimpleJson
}

type Tag struct {
	ID uint

	Title  string  `json:"title"`
	Data   TagData `gorm:"type:longtext" json:"-"`
	User   User    `json:"-"`
	UserID uint    `gorm:"column:userId" json:"-"`
	Nodes  []Node  `gorm:"" json:"-"`
}

func (Tag) TableName() string {
	return "tag"
}
