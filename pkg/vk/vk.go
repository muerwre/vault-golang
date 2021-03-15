package vk

import (
	"context"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/sirupsen/logrus"
	"io"
	"strings"
)

type Vk struct {
	api   *api.VK
	log   *logrus.Logger
	token string

	GroupId uint
}

func NewVk(token string, gid uint, log *logrus.Logger) *Vk {
	return &Vk{
		api:     api.NewVK(token),
		token:   token,
		log:     log,
		GroupId: gid,
	}
}

func (v Vk) CreatePost(ctx context.Context, msg string, url string, thumbnail *io.Reader) error {
	files := ""

	if thumbnail != nil {
		// TODO: upload photo here
		files = ""
	}

	req := api.Params{
		"message":            strings.Join([]string{msg}, " "),
		"attachment":         strings.Join([]string{files, url}, ","),
		"owner_id":           int32(v.GroupId) * -1,
		"from_group":         "1",
		"mute_notifications": "1",
	}

	_, err := v.api.WallPost(req)

	return err
}
