package request

type VkApiRequest struct {
	Response []struct {
		Id        int    `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Photo     string `json:"photo"`
	} `json:"response"`
}

type OAuthAttachConfirmRequest struct {
	Token string `json:"token"`
}
