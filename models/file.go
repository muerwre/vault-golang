package models

var FILE_TYPES = map[string]string{
	"IMAGE": "image",
	"AUDIO": "audio",
}

type FileMetadata struct {
	*SimpleJson

	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Id3title  string `json:"id3title"`
	Id3artist string `json:"id3artist"`
	Title     string `json:"title"`
	Duration  int    `json:"duration"`
}

type File struct {
	*Model

	Name     string       `json:"name"`
	OrigName string       `json:"-"`
	Path     string       `json:"-"`
	FullPath string       `json:"-"`
	Url      string       `json:"url"`
	Size     int          `json:"size"`
	Type     string       `json:"type"`
	Mime     string       `json:"mime"`
	Target   string       `json:"-"`
	User     *User        `json:"-" gorm:"foreignkey:UserID"`
	UserID   uint         `gorm:"column:userId" json:"-"`
	Metadata FileMetadata `sql:"metadata" gorm:"name:metadata;type:longtext" json:"metadata"`
	Comments []*Comment   `gorm:"many2many:comment_files_file;jointable_foreignkey:fileId;association_jointable_foreignkey:commentId" json:"-"`
	Nodes    []*Node      `gorm:"many2many:node_files_file;jointable_foreignkey:fileId;association_jointable_foreignkey:nodeId" json:"-"`
}

func (File) TableName() string {
	return "file"
}
