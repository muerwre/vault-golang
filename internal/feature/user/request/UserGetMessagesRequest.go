package request

import "time"

type UserGetMessagesRequest struct {
	Before *time.Time `json:"before" form:"before"`
	After  *time.Time `json:"after" form:"after"`
	Limit  int        `json:"limit" form:"limit"`
}

func (r *UserGetMessagesRequest) Normalize() {
	if r.Before == nil || r.Before.After(time.Now()) {
		now := time.Now()
		r.Before = &now
	}

	if r.After == nil || r.After.After(*r.Before) {
		now := time.Time{}
		r.After = &now
	}

	if r.Limit <= 0 || r.Limit > 200 {
		r.Limit = 50
	}
}
