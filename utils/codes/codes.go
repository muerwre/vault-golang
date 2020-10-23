package codes

type Code string

const (
	NotAnEmail                   string = "Not_An_Email"
	UserNotFound                 string = "User_Not_found"
	TooShirt                     string = "Is_Too_Shirt"
	EmptyRequest                 string = "Empty_Request"
	FilesRequired                string = "Files_Required"
	NodeNotFound                 string = "Node_Not_Found"
	TextRequired                 string = "Text_Required"
	UrlInvalid                   string = "Url_Invalid"
	FilesAudioRequired           string = "Files_Audio_Required"
	NotEnoughRights              string = "Not_Enough_Rights"
	IncorrectData                string = "Incorrect_Data"
	ImageConversionFailed        string = "Image_Conversion_Failed"
	UserExist                    string = "User_Exist"
	IncorrectPassword            string = "Incorrect_Password"
	CodeIsInvalid                string = "Code_Is_Invalid"
	Required                     string = "Required"
	CommentNotFound              string = "Comment_Not_Found"
	IncorrectType                string = "Incorrect_Node_Type"
	UnexpectedBehavior           string = "Unexpected_Behavior"
	UnknownFileType              string = "Unknown_File_Type"
	FilesIsTooBig                string = "File_Is_Too_Big"
	OAuthCodeIsEmpty             string = "OAuth_Code_Is_Empty"
	OAuthUnknownProvider         string = "OAuth_Unknown_Provider"
	OAuthInvalidData             string = "OAuth_Invalid_Data"
	OAuthConflict                string = "OAuth_Conflict"
	UsernameIsShort              string = "Username_Is_Short"
	UsernameContainsInvalidChars string = "Username_Contains_Invalid_Chars"
	PasswordIsShort              string = "Password_Is_Short"
	UserExistWithEmail           string = "User_Exist_With_Email"
	UserExistWithSocial          string = "User_Exist_With_Social"
	UserExistWithUsername        string = "User_Exist_With_Username"
	CantSaveComment              string = "CantSaveComment"
	UnknownNodeType              string = "UnknownNodeType"
	CantSaveNode                 string = "CantSaveNode"
	CantLoadUser                 string = "CantLoadUser"
	InputTooShirt                string = "InputTooShirt"
	CantSaveUser                 string = "CantSaveUser"
	CantDeleteComment            string = "CantDeleteComment"
	CantRestoreComment           string = "CantRestoreComment"
	MessageNotFound              string = "MessageNotFound"
	CommentTooLong               string = "CommentTooLong"
)

var ValidationToCode = map[string]string{
	"gte":      InputTooShirt,
	"email":    NotAnEmail,
	"required": Required,
}
