package models

type Social struct {
	ID           uint   `json:"-"`
	Provider     string `json:"provider"`
	AccountId    string `gorm:"column:account_id" json:"id"`
	AccountName  string `json:"name"`
	AccountPhoto string `json:"photo"`

	User   *User `json:"-" gorm:"foreignkey:UserID"`
	UserID *uint `gorm:"column:userId" json:"-"`
}

func (Social) TableName() string {
	return "social"
}
