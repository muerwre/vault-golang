package controllers

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/utils"
	"github.com/muerwre/vault-golang/utils/codes"
	"github.com/sirupsen/logrus"
	"net/http"
)

type OAuthController struct {
	Config app.Config
	DB     db.DB

	credentials utils.OAuthCredentials
}

func (oc *OAuthController) Init() {
	oc.credentials = utils.OAuthCredentials{
		VkClientId:         oc.Config.VkClientId,
		VkClientSecret:     oc.Config.VkClientSecret,
		VkCallbackUrl:      oc.Config.VkCallbackUrl,
		GoogleClientId:     oc.Config.GoogleClientId,
		GoogleClientSecret: oc.Config.GoogleClientSecret,
		GoogleCallbackUrl:  oc.Config.GoogleCallbackUrl,
	}
}

func (oc OAuthController) ProviderMiddleware(c *gin.Context) {
	provider := c.Param("provider")
	current, err := utils.OAuthConfigs.GetByName(provider)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.Set("Provider", current)
}

func (oc OAuthController) Redirect(c *gin.Context) {
	provider := c.MustGet("Provider").(*utils.OAuthConfig)
	target := c.Param("target")
	config := provider.ConfigCreator(oc.credentials, target)
	c.Redirect(http.StatusTemporaryRedirect, config.AuthCodeURL("pseudo-random")) // TODO: pseudo-random can be in payload
	return
}

func (oc OAuthController) Process(target string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		provider := c.MustGet("Provider").(*utils.OAuthConfig)
		code := c.Query("code")

		if code == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": codes.OAuthCodeIsEmpty})
			return
		}

		config := provider.ConfigCreator(oc.credentials, target)
		token, err := config.Exchange(ctx, code)

		if err != nil {
			logrus.Warnf("Failed to get token: %v", err.Error())
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Failed to get token"})
			return
		}

		data, err := provider.Parser(token)

		if err != nil {
			logrus.Warnf("Failed to get token: %v", err.Error())
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}

		data.Fetched, err = provider.Fetcher(data.Token)

		if err != nil {
			logrus.Warnf("Failed to get token: %v", err.Error())
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}

		c.Set("UserData", data)
		c.Next()
	}
}

func (oc OAuthController) Attach(c *gin.Context) {
	ud := c.MustGet("UserData").(*utils.OauthUserData)

	if _, err := oc.DB.SocialRepository.FindOne(ud.Provider, ud.Id); err == nil {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": codes.UserExist})
		return
	}

	// TODO: create jwt token with attach pair

	c.AbortWithStatusJSON(http.StatusOK, gin.H{"data": ud})
	return
}

func (oc OAuthController) Login(c *gin.Context) {
	ud := c.MustGet("UserData").(*utils.OauthUserData)

	social, err := oc.DB.SocialRepository.FindOne(ud.Provider, ud.Id)

	if err == nil {
		token := oc.DB.UserRepository.GenerateTokenFor(social.User)
		c.AbortWithStatusJSON(http.StatusOK, gin.H{"token": token.Token})
		return
	}

	// TODO: create user
	// TODO: upload photo
	// TODO: create social
	// TODO: generate token

	c.String(http.StatusOK, "TODO:")
	return
}
