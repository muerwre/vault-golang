package controllers

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/models"
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

// ProviderMiddleware generates Provider context by :provider url param
func (oc OAuthController) ProviderMiddleware(c *gin.Context) {
	provider := c.Param("provider")
	current, err := utils.OAuthConfigs.GetByName(provider)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.Set("Provider", current)
}

// Redirect redirects user to oauth endpoint, that redirects back to /process/:target?code= urls
func (oc OAuthController) Redirect(c *gin.Context) {
	provider := c.MustGet("Provider").(*utils.OAuthConfig)
	target := c.Param("target")
	config := provider.ConfigCreator(oc.credentials, target)
	c.Redirect(http.StatusTemporaryRedirect, config.AuthCodeURL("pseudo-random")) // TODO: pseudo-random can be in payload
	return
}

// Process is a middleware, that fetches oauth data from provider and passes it further
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

// Attach gets fetched from oauth data and encodes it as JWT to send back to frontend, so it can call AttachConfirm with it
func (oc OAuthController) Attach(c *gin.Context) {
	ud := c.MustGet("UserData").(*utils.OauthUserData)
	claim := new(utils.OauthUserDataClaim).Init(*ud)
	token, err := utils.EncodeJwtToken(claim)

	if err != nil {
		logrus.Warnf("Failed to create attach token: %v", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.AbortWithStatusJSON(http.StatusOK, gin.H{"token": token})
}

// AttachConfirm gets user oauth data from token and creates social connection for it
func (oc OAuthController) AttachConfirm(c *gin.Context) {
	u := c.MustGet("User").(*models.User)
	claim, err := utils.DecodeOauthClaimFromRequest(c)

	if err != nil {
		logrus.Warnf("Failed to perform attach confirm: %v", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": codes.OAuthInvalidData})
		return
	}

	if exist, err := oc.DB.SocialRepository.FindOne(claim.Data.Provider, claim.Data.Id); err == nil {
		// User already has this social account
		if exist.User.ID == u.ID {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"social": exist})
			return
		}

		// Another user has it
		c.AbortWithStatusJSON(http.StatusOK, gin.H{"error": codes.UserExist})
		return
	}

	social := &models.Social{
		Provider:     claim.Data.Provider,
		AccountId:    claim.Data.Id,
		AccountPhoto: claim.Data.Fetched.Photo,
		AccountName:  claim.Data.Fetched.Name,
		User:         u,
	}

	oc.DB.SocialRepository.Create(social)

	c.AbortWithStatusJSON(http.StatusOK, social)
}

// Login logs user in or registers account
func (oc OAuthController) Login(c *gin.Context) {
	ud := c.MustGet("UserData").(*utils.OauthUserData)

	social, err := oc.DB.SocialRepository.FindOne(ud.Provider, ud.Id)

	// Social exist, login user
	if err == nil {
		token := oc.DB.UserRepository.GenerateTokenFor(social.User)
		// TODO: update social info here
		c.AbortWithStatusJSON(http.StatusOK, gin.H{"token": token.Token})
		return
	}

	claim := new(utils.OauthUserDataClaim).Init(*ud)
	token, err := utils.EncodeJwtToken(claim)

	c.JSON(http.StatusOK, gin.H{"token": token})
	return
}

func (oc OAuthController) Register(c *gin.Context) {
	claim, err := utils.DecodeOauthClaimFromRequest(c)

	if err != nil {
		logrus.Warnf("Failed to perform login confirm: %v", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": codes.OAuthInvalidData})
		return
	}

	// Check if there's no account with this email
	if _, err := oc.DB.UserRepository.GetByEmail(claim.Data.Email); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": codes.UserExist})
		return
	}

	// Check if any user has this social
	if _, err := oc.DB.SocialRepository.FindOne(claim.Data.Provider, claim.Data.Id); err == nil {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{"error": codes.UserExist})
		return
	}

	// TODO: upload photo
	// TODO: create user
	// TODO: create social
	// TODO: generate token
}
