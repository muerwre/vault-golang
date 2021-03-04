package request

import (
	"regexp"
	"strings"
)

type SearchNodeRequest struct {
	Text string `json:"text" form:"text"`
	Skip int    `json:"skip" form:"skip"`
	Take int    `json:"take" form:"take"`
}

func (r *SearchNodeRequest) Sanitize() {
	re := regexp.MustCompile(`[^А-Яа-я\w\s-.,\!\?]+`)
	r.Text = strings.TrimSpace(re.ReplaceAllString(r.Text, ""))

	if r.Take <= 0 {
		r.Take = 20
	}

	if r.Skip < 0 {
		r.Skip = 0
	}
}
