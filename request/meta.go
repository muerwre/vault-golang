package request

type MetaYoutubeRequest struct {
	Items []struct {
		Id      string `json:"id"`
		Snippet struct {
			Title string `json:"title"`
		} `json:"snippet"`
	} `json:"items"`
}
