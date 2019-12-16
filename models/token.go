package models

type Token struct {
	*Model

	Token string `json:"-"`

	User   *User `json:"-"`
	UserID uint  `gorm:"column:userId" json:"-"`
}

func (Token) TableName() string {
	return "token"
}
