package models

import (
	"database/sql/driver"
	"encoding/json"
)

type FileTypesStruct struct {
	IMAGE string
	AUDIO string
}

var FileTypes = FileTypesStruct{
	IMAGE: "image",
	AUDIO: "audio",
}

type FileMimeStruct struct {
	JPEG  string
	PNG   string
	GIF   string
	MPEG3 string
	MPEG  string
	MP3   string
}

var FileMime = FileMimeStruct{
	JPEG:  "image/jpeg",
	PNG:   "image/png",
	GIF:   "image/gif",
	MPEG3: "audio/mpeg3",
	MPEG:  "audio/mpeg",
	MP3:   "audio/mp3",
}

type FileTypeToMimeStruct map[string][]string

var FileTypeToMime = FileTypeToMimeStruct{
	FileTypes.IMAGE: {FileMime.JPEG, FileMime.PNG, FileMime.GIF},
	FileTypes.AUDIO: {FileMime.MP3, FileMime.MPEG, FileMime.MPEG3},
}

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
