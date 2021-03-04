package models

import "time"

type NotificationSettings struct {
	ID                    uint
	User                  *User `json:"user" gorm:"foreignkey:UserID"`
	UserID                *uint `gorm:"column:userId;uniqueIndex:user_id" json:"-"`
	LastSeenNotifications time.Time
	SubscribedToFlow      bool `json:"subscribed_to_flow"`
	SubscribedToComments  bool `json:"subscribed_to_comments" gorm:"default:true"`
}

func (NotificationSettings) TableName() string {
	return "notification_settings"
}
