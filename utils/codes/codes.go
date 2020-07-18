package codes

type Code string

const (
	NotAnEmail            string = "Not_An_Email"
	UserNotFound          string = "User_Not_found"
	TooShirt              string = "Is_Too_Shirt"
	EmptyRequest          string = "Empty_Request"
	FilesRequired         string = "Files_Required"
	NodeNotFound          string = "Node_Not_Found"
	TextRequired          string = "Text_Required"
	UrlInvalid            string = "Url_Invalid"
	FilesAudioRequired    string = "Files_Audio_Required"
	NotEnoughRights       string = "Not_Enough_Rights"
	IncorrectData         string = "Incorrect_Data"
	ImageConversionFailed string = "Image_Conversion_Failed"
	UserExist             string = "User_Exist"
	IncorrectPassword     string = "Incorrect_Password"
	CodeIsInvalid         string = "Code_Is_Invalid"
	Required              string = "Required"
	CommentNotFound       string = "Comment_Not_Found"
	IncorrectType         string = "Incorrect_Node_Type"
	UnexpectedBehavior    string = "Unexpected_Behavior"
	UnknownFileType       string = "Unknown_File_Type"
	FilesIsTooBig         string = "File_Is_Too_Big"
	OAuthCodeIsEmpty      string = "OAuth_Code_Is_Empty"
	OAuthUnknownProvider  string = "OAuth_Unknown_Provider"
	OAuthInvalidData      string = "OAuth_Invalid_Data"
	OAuthConflict         string = "OAuth_Conflict"
)

var ValidationToCode = map[string]string{
	"gte":      TooShirt,
	"email":    NotAnEmail,
	"required": Required,
}
