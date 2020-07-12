package models

type Social struct {
	ID        uint
	Provider  string `json:"-"`
	AccountId string `gorm:"column:account_id" json:"-"`
	User      *User  `json:"-" gorm:"foreignkey:UserID"`
	UserID    *uint  `gorm:"column:userId" json:"-"`
}

func (Social) TableName() string {
	return "social"
}
