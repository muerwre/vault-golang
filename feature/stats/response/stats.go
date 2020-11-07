package response

type StatsUsers struct {
	Total int `json:"total"`
	Alive int `json:"alive"`
}

type StatsNodes struct {
	Images int `json:"images"`
	Audios int `json:"audios"`
	Videos int `json:"videos"`
	Texts  int `json:"texts"`
	Total  int `json:"total"`
}

type StatsComments struct {
	Total int `json:"total"`
}

type StatsFiles struct {
	Count int `json:"count"`
	Size  int `json:"size"`
}

type StatsTimestamps struct {
	BorisLastComment string `json:"boris_last_comment"`
	FlowLastPost     string `json:"flow_last_post"`
}

type StatsResponse struct {
	StatsUsers      `json:"users"`
	StatsNodes      `json:"nodes"`
	StatsComments   `json:"comments"`
	StatsFiles      `json:"files"`
	StatsTimestamps `json:"timestamps"`
}
