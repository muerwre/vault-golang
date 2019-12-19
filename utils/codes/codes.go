package codes

type Code string

const (
	NOT_AN_EMAIL            string = "Not_An_Email"
	USER_NOT_FOUND          string = "User_Not_found"
	TOO_SHIRT               string = "Is_Too_Shirt"
	EMPTY_REQUEST           string = "Empty_Request"
	FILES_REQUIRED          string = "Files_Required"
	NODE_NOT_FOUND          string = "Node_Not_Found"
	TEXT_REQUIRED           string = "Text_Required"
	URL_INVALID             string = "Url_Invalid"
	FILES_AUDIO_REQUIRED    string = "Files_Audio_Required"
	NOT_ENOUGH_RIGHTS       string = "Not_Enough_Rights"
	INCORRECT_DATA          string = "Incorrect_Data"
	IMAGE_CONVERSION_FAILED string = "Image_Conversion_Failed"
	USER_EXIST              string = "User_Exist"
	INCORRECT_PASSWORD      string = "Incorrect_Password"
	CODE_IS_INVALID         string = "Code_Is_Invalid"
	REQUIRED                string = "Required"
	COMMENT_NOT_FOUND       string = "Comment_Not_Found"
)

var VALIDATION_TO_CODE = map[string]string{
	"gte":      TOO_SHIRT,
	"email":    NOT_AN_EMAIL,
	"required": REQUIRED,
}
