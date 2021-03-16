package controller

type VkNotificationsConfig struct {
	Enabled        bool
	GroupId        uint
	ApiKey         string
	Delay          uint
	CooldownMins   uint
	PurgeAfterDays uint
	UrlPrefix      string
	UploadPath     string
}
