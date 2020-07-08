package models

import "time"

type Message struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `sql:"index" json:"-"`

	Text string `json:"text"`

	FilesOrder CommaUintArray `gorm:"column:files_order;type:longtext;" json:"files_order"`
	Files      []*File        `gorm:"many2many:message_files_file;jointable_foreignkey:messageId;association_jointable_foreignkey:fileId" json:"files"`

	To   *User `json:"to" gorm:"foreignkey:ToID"`
	ToID uint  `gorm:"column:toId" json:"-"`

	From   *User `json:"from" gorm:"foreignkey:FromID"`
	FromID uint  `gorm:"column:fromId" json:"-"`
}

func (Message) TableName() string {
	return "message"
}

func (m Message) IsValid() bool {
	return len(m.Text) >= 1
}
