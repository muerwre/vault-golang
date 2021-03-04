package usecase

import (
	"context"
	"github.com/muerwre/vault-golang/internal/app"
	"github.com/muerwre/vault-golang/internal/db"
	"github.com/muerwre/vault-golang/internal/db/models"
	repository2 "github.com/muerwre/vault-golang/internal/db/repository"
	utils2 "github.com/muerwre/vault-golang/internal/feature/oauth/utils"
)

type OauthUsecase struct {
	credentials utils2.OAuthCredentials
	oauth       repository2.OauthRepository
}

func (ou *OauthUsecase) Init(db db.DB, config app.Config) *OauthUsecase {
	ou.credentials = utils2.OAuthCredentials{
		VkClientId:         config.VkClientId,
		VkClientSecret:     config.VkClientSecret,
		VkCallbackUrl:      config.VkCallbackUrl,
		GoogleClientId:     config.GoogleClientId,
		GoogleClientSecret: config.GoogleClientSecret,
		GoogleCallbackUrl:  config.GoogleCallbackUrl,
	}
	ou.oauth = *db.Social
	return ou
}

func (ou OauthUsecase) GetRedirectUrlForProvider(provider *utils2.OAuthConfig) string {
	config := provider.ConfigCreator(ou.credentials)
	return config.AuthCodeURL("pseudo-random")
}

func (ou OauthUsecase) GetTokenData(provider *utils2.OAuthConfig, code string) (*utils2.OauthUserData, error) {
	ctx := context.Background()
	config := provider.ConfigCreator(ou.credentials)
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	data, err := provider.Parser(token)
	if err != nil {
		return nil, err
	}

	data.Fetched, err = provider.Fetcher(data.Token)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (ou OauthUsecase) GetSocialById(provider string, id string) (*models.Social, error) {
	return ou.oauth.FindOne(provider, id)
}

func (ou OauthUsecase) CreateSocialFromClaim(claim utils2.OauthUserDataClaim, u *models.User) (*models.Social, error) {
	social := &models.Social{
		Provider:     claim.Data.Provider,
		AccountId:    claim.Data.Id,
		AccountPhoto: claim.Data.Fetched.Photo,
		AccountName:  claim.Data.Fetched.Name,
		User:         u,
	}

	if err := ou.oauth.Create(social); err != nil {
		return nil, err
	}

	return social, nil
}

func (ou OauthUsecase) GetSocialsOfUser(user *models.User) ([]*models.Social, error) {
	return ou.oauth.OfUser(user.ID)
}

func (ou OauthUsecase) DeleteSocialByUserProviderAndId(user *models.User, provider string, id string) error {
	return ou.oauth.DeleteOfUser(user.ID, provider, id)
}
