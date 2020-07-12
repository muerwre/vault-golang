package models

import (
	"database/sql/driver"
	"encoding/json"
)

const (
	FileTypeImage string = "image"
	FileTypeAudio string = "audio"
)

type FileMetadata struct {
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
	Metadata FileMetadata `sql:"metadata" gorm:"column:metadata;type:longtext" json:"metadata"`
	Comments []*Comment   `gorm:"many2many:comment_files_file;jointable_foreignkey:fileId;association_jointable_foreignkey:commentId" json:"-"`
	Nodes    []*Node      `gorm:"many2many:node_files_file;jointable_foreignkey:fileId;association_jointable_foreignkey:nodeId" json:"-"`
}

func (File) TableName() string {
	return "file"
}

func (s *FileMetadata) Scan(src interface{}) error {
	return json.Unmarshal(src.([]byte), &s)
}

func (s FileMetadata) Value() (driver.Value, error) {
	val, err := json.Marshal(s)
	return string(val), err
}

const (
	FileMimeJpeg  string = "image/jpeg"
	FileMimePng   string = "image/png"
	FileMimeGif   string = "image/gif"
	FileMimeMpeg3 string = "audio/mpeg3"
	FileMimeMpeg  string = "audio/mpeg"
	FileMimeMp3   string = "audio/mp3"
)

var FileTypeToMime = map[string][]string{
	FileTypeImage: {FileMimeJpeg, FileMimePng, FileMimeGif},
	FileTypeAudio: {FileMimeMp3, FileMimeMpeg, FileMimeMpeg3},
}

func FileGetTypeByMime(fileMime string) string {
	for fileType, mimes := range FileTypeToMime {
		for _, mimeType := range mimes {
			if mimeType == fileMime {
				return fileType
			}
		}
	}

	return ""
}

const (
	FileTargetNodes    string = "nodes"
	FileTargetComments string = "comments"
	FileTargetProfiles string = "profiles"
	FileTargetOther    string = "other"
)

var FileTargets = []string{FileTargetComments, FileTargetNodes, FileTargetProfiles, FileTargetOther}

func FileValidateTarget(target string) bool {
	for _, v := range FileTargets {
		if target == v {
			return true
		}
	}

	return false
}
