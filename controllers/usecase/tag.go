package usecase

import (
	"fmt"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/response"
	"net/url"
	"strconv"
)

type TagUsecase struct {
	db   db.DB
	conf app.Config
}

func (uc *TagUsecase) Init(db db.DB, conf app.Config) *TagUsecase {
	uc.db = db
	uc.conf = conf
	return uc
}

func (uc TagUsecase) GetTagByName(name string) (tag *models.Tag, err error) {
	if name == "" {
		return nil, fmt.Errorf("attempting to fetch empty tag")
	}

	return uc.db.Tag.GetByName(name)
}

func (uc TagUsecase) GetNodesOfTag(tag models.Tag, limit string, offset string) ([]response.SearchNodeResponseNode, int, error) {
	l, err := strconv.Atoi(limit)
	if err != nil || l <= 0 {
		l = 20
	}

	o, err := strconv.Atoi(offset)
	if err != nil || o < 0 {
		o = 0
	}

	nodes, count, err := uc.db.Tag.GetNodesOfTag(tag, l, o)
	if err != nil {
		return nil, 0, err
	}

	results := make([]response.SearchNodeResponseNode, len(nodes))
	for k, v := range nodes {
		results[k] = *new(response.SearchNodeResponseNode).Init(*v)
	}

	return results, count, nil
}

func (uc TagUsecase) GetTagsForAutocomplete(s string) ([]string, error) {
	search, err := url.QueryUnescape(s)

	if err != nil || search == "" {
		return nil, nil
	}

	tags, err := uc.db.Tag.GetLike(search)

	if err != nil {
		return nil, err
	}

	res := make([]string, len(tags))

	for k, v := range tags {
		res[k] = v.Title
	}

	return res, nil
}
