package request

import "time"

type LabListQuery struct {
	After *time.Time `form:"after"`
	Limit int        `form:"limit"`
}
