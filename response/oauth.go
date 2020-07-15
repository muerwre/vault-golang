package response

type OAuthResultResponseUser struct {
	Name string `json:"name"`
}

type OAuthResultResponse struct {
	Token string                  `json:"token"`
	User  OAuthResultResponseUser `json:"user"`
}
