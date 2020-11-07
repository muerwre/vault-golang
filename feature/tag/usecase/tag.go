package tagUsecase

import (
	"fmt"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/db"
	response2 "github.com/muerwre/vault-golang/feature/search/response"
	"github.com/muerwre/vault-golang/feature/tag/utils"
	"github.com/muerwre/vault-golang/models"
	"net/url"
	"strconv"
	"strings"
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

func (uc TagUsecase) GetNodesOfTag(tag models.Tag, limit string, offset string) ([]response2.SearchNodeResponseNode, int, error) {
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

	results := make([]response2.SearchNodeResponseNode, len(nodes))
	for k, v := range nodes {
		results[k] = *new(response2.SearchNodeResponseNode).Init(*v)
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

func (uc TagUsecase) FindOrCreateTags(titles []string) ([]*models.Tag, error) {
	// make incoming tags lowercase
	for i := 0; i < len(titles); i += 1 {
		titles[i] = strings.ToLower(titles[i])
	}

	if len(titles) == 0 {
		return make([]*models.Tag, 0), nil
	}

	// load tags
	tags, err := uc.db.Tag.FindTagsByTitleList(titles)
	if err != nil {
		return nil, err
	}

	// create missed tags
	for _, v := range titles {
		if !utils.TagArrayContains(tags, v) && len(v) > 0 {
			if tag, err := uc.db.Tag.CreateTagFromTitle(v); err == nil {
				tags = append(tags, tag)
			}
		}
	}

	return tags, err
}
