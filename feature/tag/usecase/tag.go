package tagUsecase

import (
	"fmt"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/db/models"
	repository2 "github.com/muerwre/vault-golang/db/repository"
	response2 "github.com/muerwre/vault-golang/feature/search/response"
	"github.com/muerwre/vault-golang/feature/tag/utils"
	"net/url"
	"strconv"
	"strings"
)

type TagUsecase struct {
	tag repository2.TagRepository
}

func (uc *TagUsecase) Init(db db.DB) *TagUsecase {
	uc.tag = *db.Tag
	return uc
}

func (uc TagUsecase) GetTagByName(name string) (tag *models.Tag, err error) {
	if name == "" {
		return nil, fmt.Errorf("attempting to fetch empty tag")
	}

	return uc.tag.GetByName(name)
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

	nodes, count, err := uc.tag.GetNodesOfTag(tag, l, o)
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

	tags, err := uc.tag.GetLike(search)

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
	tags, err := uc.tag.FindTagsByTitleList(titles)
	if err != nil {
		return nil, err
	}

	// create missed tags
	for _, v := range titles {
		if !utils.TagArrayContains(tags, v) && len(v) > 0 {
			if tag, err := uc.tag.CreateTagFromTitle(v); err == nil {
				tags = append(tags, tag)
			}
		}
	}

	return tags, err
}
