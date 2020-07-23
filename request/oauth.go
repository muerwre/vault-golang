package request

import (
	"fmt"
	"github.com/muerwre/vault-golang/utils/codes"
)

type VkApiRequest struct {
	Response []struct {
		Id        int    `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Photo     string `json:"photo"`
	} `json:"response"`
}

type OAuthRegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"string"`
}

func (req *OAuthRegisterRequest) Valid() error {
	switch {
	case len(req.Username) < 2:
		return fmt.Errorf(codes.UsernameIsShort)
	case len(req.Password) < 6:
		return fmt.Errorf(codes.PasswordIsShort)
	default:
		return nil
	}
}
