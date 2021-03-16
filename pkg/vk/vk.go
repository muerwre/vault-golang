package vk

import (
	"context"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
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

func (v Vk) CreatePost(ctx context.Context, msg string, u string, thumbnail string) error {
	files := ""

	if thumbnail != "" {
		f, err := os.Open(thumbnail)
		if err == nil {
			resp, err := v.api.UploadPhotoGroup(int(v.GroupId), 166682397, f)
			if err == nil && len(resp) > 0 {
				files = getPhotoSource(resp[0], v.GroupId)
			}
		}

	}

	req := api.Params{
		"message":            strings.Join([]string{msg}, " "),
		"attachments":        strings.Join([]string{files}, ","),
		"owner_id":           int32(v.GroupId) * -1,
		"from_group":         "1",
		"mute_notifications": "1",
	}

	resp, err := v.api.WallPost(req)

	v.log.Infof("%+v, %s", resp, u)

	return err
}

func getPhotoSource(resp object.PhotosPhoto, gid uint) string {
	return strings.Join([]string{
		"photo",
		strconv.Itoa(resp.OwnerID),
		"_",
		strconv.Itoa(resp.ID),
	}, "")
}
