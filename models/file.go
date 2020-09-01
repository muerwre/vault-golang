package models

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/muerwre/vault-golang/constants"
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
	Path     string       `json:"path"`
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

func FileGetTypeByMime(fileMime string) string {
	for fileType, mimes := range constants.FileTypeToMime {
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

// FileHasInvalidMetatada reports if file metadata is empty
func (f File) FileHasInvalidMetatada() bool {
	return f.Type == constants.FileTypeImage && (f.Metadata.Width == 0 && f.Metadata.Height == 0)
}
