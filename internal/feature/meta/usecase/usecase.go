package usecase

import (
	"github.com/muerwre/vault-golang/internal/app"
	"github.com/muerwre/vault-golang/internal/db"
	"github.com/muerwre/vault-golang/internal/db/models"
	repository2 "github.com/muerwre/vault-golang/internal/db/repository"
	"github.com/muerwre/vault-golang/internal/service/google"
)

type MetaUsecase struct {
	meta    repository2.MetaRepository
	youtube google.YoutubeService
}

func (mu *MetaUsecase) Init(db db.DB, config app.Config) *MetaUsecase {
	mu.youtube = *new(google.YoutubeService).Init(config.GoogleApiKey)
	mu.meta = *db.Meta
	return mu
}

func (mu MetaUsecase) FetchYoutubeInfoForIds(ids []string) (map[string]*models.Embed, error) {
	return mu.youtube.FetchYoutubeInfoForIds(ids)
}

func (mu MetaUsecase) GetEmbedsFromDbOrFetchFromGoogle(ids []string) (map[string]*models.Embed, error) {
	if len(ids) < 1 {
		return make(map[string]*models.Embed, 0), nil
	}

	embeds, err := mu.meta.GetEmbedsById(ids, "youtube")

	lost := make([]string, 0)

	for _, v := range ids {
		if embeds[v] == nil {
			lost = append(lost, v)
		}
	}

	if len(lost) > 0 {
		created, _ := mu.FetchYoutubeInfoForIds(lost)

		if len(created) > 0 {
			values := make([]models.Embed, len(created))
			i := 0

			for k, v := range created {
				values[i] = *v
				embeds[k] = v
				i++
			}

			_ = mu.meta.SaveEmbeds(values)
		}
	}

	return embeds, err
}
