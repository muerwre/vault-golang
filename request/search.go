package request

type SearchNodeRequest struct {
	Text string `json:"text" form:"text"`
	Skip int    `json:"skip" form:"skip"`
	Take int    `json:"take" form:"take"`
}
