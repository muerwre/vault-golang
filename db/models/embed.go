package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type EmbedMetadata struct {
	Title    string `json:"title"`
	Thumb    string `json:"thumb"`
	Duration string `json:"duration"`
}

type Embed struct {
	ID       uint          `gorm:"primary_key" json:"id"`
	Provider string        `json:"provider"`
	Address  string        `json:"address"`
	Metadata EmbedMetadata `json:"metadata" gorm:"name:metadata;type:longtext" json:"route"`

	CreatedAt time.Time  `json:"created_at" json:"-"`
	UpdatedAt time.Time  `json:"updated_at" json:"-"`
	DeletedAt *time.Time `sql:"index" json:"-"`
}

func (Embed) TableName() string {
	return "embed"
}

func (p *EmbedMetadata) Scan(src interface{}) error {
	return json.Unmarshal(src.([]byte), &p)
}

func (p EmbedMetadata) Value() (driver.Value, error) {
	val, err := json.Marshal(p)
	return string(val), err
}
