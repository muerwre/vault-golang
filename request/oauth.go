package request

import (
	"fmt"
	"github.com/muerwre/vault-golang/constants"
	"github.com/muerwre/vault-golang/utils/codes"
	"regexp"
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

func (req *OAuthRegisterRequest) Valid() (map[string]error, error) {
	errors := map[string]error{}
	usernameRegexp := regexp.MustCompile(constants.UsernameRegexp)

	switch {
	case len(req.Username) < 2:
		errors["username"] = fmt.Errorf(codes.UsernameIsShort)
		fallthrough
	case !usernameRegexp.MatchString(req.Username):
		errors["username"] = fmt.Errorf(codes.UsernameContainsInvalidChars)
		fallthrough
	case len(req.Password) < 6:
		errors["password"] = fmt.Errorf(codes.PasswordIsShort)
		fallthrough
	default:
	}

	if len(errors) == 0 {
		return nil, nil
	}

	return errors, fmt.Errorf(codes.IncorrectData)
}
