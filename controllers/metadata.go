package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/request"
	"github.com/muerwre/vault-golang/utils/codes"
	"io/ioutil"
	"net/http"
	"strings"
)

type MetaController struct {
	Config app.Config
	DB     db.DB
}

func (mc MetaController) FetchYoutubeInfo(ids []string) (map[string]*models.Embed, error) {
	req, err := http.NewRequest(http.MethodGet, "https://www.googleapis.com/youtube/v3/videos", nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("key", mc.Config.GoogleApiKey)
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

func (mc MetaController) GetYoutubeTitles(c *gin.Context) {
	ids := strings.Split(c.Query("ids"), ",")

	if len(ids) < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.EmptyRequest})
	}

	embeds, _ := mc.DB.Meta.GetEmbedsById(ids, "youtube")

	lost := make([]string, 0)

	for _, v := range ids {
		if embeds[v] == nil {
			lost = append(lost, v)
		}
	}

	if len(lost) > 0 {
		created, _ := mc.FetchYoutubeInfo(lost)

		if len(created) > 0 {
			values := make([]models.Embed, len(created))
			i := 0

			for k, v := range created {
				values[i] = *v
				embeds[k] = v
				i++
			}

			mc.DB.Meta.SaveEmbeds(values)
		}
	}

	c.JSON(http.StatusOK, gin.H{"items": embeds})
}
