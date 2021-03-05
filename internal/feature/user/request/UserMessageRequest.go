package request

type UserMessageRequestItem struct {
	ID   uint   `json:"id"`
	Text string `json:"text"`
}

type UserMessageRequest struct {
	Message UserMessageRequestItem `json:"message"`
}
