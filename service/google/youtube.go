package google

import (
	"encoding/json"
	"github.com/muerwre/vault-golang/db/models"
	"github.com/muerwre/vault-golang/feature/meta/request"
	"io/ioutil"
	"net/http"
	"strings"
)

type YoutubeService struct {
	key string
}

func (ys *YoutubeService) Init(key string) *YoutubeService {
	ys.key = key
	return ys
}

func (ys YoutubeService) FetchYoutubeInfoForIds(ids []string) (map[string]*models.Embed, error) {
	req, err := http.NewRequest(http.MethodGet, "https://www.googleapis.com/youtube/v3/videos", nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("key", ys.key)
	q.Add("id", strings.Join(ids, ","))
	q.Add("part", "snippet")
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	respBody, _ := ioutil.ReadAll(resp.Body)

	items := &request.MetaYoutubeRequest{}
	err = json.Unmarshal(respBody, &items)

	if err != nil {
		return nil, err
	}

	result := make(map[string]*models.Embed, len(items.Items))

	for _, v := range items.Items {
		result[v.Id] = &models.Embed{
			Address:  v.Id,
			Provider: "youtube",
			Metadata: models.EmbedMetadata{
				Title: v.Snippet.Title,
			},
		}
	}

	return result, nil
}
