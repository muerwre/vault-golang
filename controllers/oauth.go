package controllers

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/muerwre/vault-golang/app"
	"github.com/muerwre/vault-golang/constants"
	"github.com/muerwre/vault-golang/db"
	"github.com/muerwre/vault-golang/models"
	"github.com/muerwre/vault-golang/request"
	"github.com/muerwre/vault-golang/utils"
	"github.com/muerwre/vault-golang/utils/codes"
	"github.com/muerwre/vault-golang/utils/passwords"
	"github.com/sirupsen/logrus"
	"net/http"
)

const (
	eventTypeProcessed string = "oauth_processed"
	eventTypeError     string = "oauth_error"
)

type OAuthController struct {
	Config app.Config
	DB     db.DB

	credentials    utils.OAuthCredentials
	fileController FileController
}

// TODO: reply to errors via c.HTML in endpoints, which opened in modals

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
	config := provider.ConfigCreator(oc.credentials)
	c.Redirect(http.StatusTemporaryRedirect, config.AuthCodeURL("pseudo-random"))
	return
}

// GetRedirectData is a middleware, that fetches oauth data from provider and passes it further
func (oc OAuthController) GetRedirectData(target string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		provider := c.MustGet("Provider").(*utils.OAuthConfig)
		code := c.Query("code")

		if code == "" {
			utils.ReplyHtmlEventWithError(c, eventTypeError, codes.OAuthCodeIsEmpty)
			c.Abort()
			return
		}

		config := provider.ConfigCreator(oc.credentials)
		token, err := config.Exchange(ctx, code)

		if err != nil {
			logrus.Warnf("Failed to get token: %v", err.Error())
			utils.ReplyHtmlEventWithError(c, eventTypeError, codes.OAuthInvalidData)
			c.Abort()
			return
		}

		data, err := provider.Parser(token)

		if err != nil {
			logrus.Warnf("Failed to get token: %v", err.Error())
			utils.ReplyHtmlEventWithError(c, eventTypeError, err.Error())
			c.Abort()
			return
		}

		data.Fetched, err = provider.Fetcher(data.Token)

		if err != nil {
			logrus.Warnf("Failed to get token: %v", err.Error())
			utils.ReplyHtmlEventWithError(c, eventTypeError, err.Error())
			c.Abort()
			return
		}

		c.Set("UserData", data)
		c.Next()
	}
}

// Process gets fetched from oauth data and encodes it as JWT to send back to frontend, so it can call AttachConfirm with it
func (oc OAuthController) Process(c *gin.Context) {
	ud := c.MustGet("UserData").(*utils.OauthUserData)
	claim := new(utils.OauthUserDataClaim).Init(*ud)
	token, err := utils.EncodeJwtToken(claim)

	if err != nil {
		logrus.Warnf("Failed to create attach token: %v", err.Error())
		utils.ReplyHtmlEventWithError(c, eventTypeError, codes.OAuthInvalidData)
		return
	}

	utils.ReplytHtmlEventWithToken(c, eventTypeProcessed, token)
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
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"account": exist})
			return
		}

		// Another user has it
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": codes.UserExistWithSocial})
		return
	}

	// Another user has it
	if user, err := oc.DB.UserRepository.GetByEmail(claim.Data.Email); err == nil && user.ID != u.ID {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": codes.UserExistWithEmail})
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

	c.AbortWithStatusJSON(http.StatusOK, gin.H{"account": social})
}

func (oc OAuthController) Login(c *gin.Context) {
	claim, err := utils.DecodeOauthClaimFromRequest(c)

	if err != nil {
		logrus.Warnf("Failed to perform login: %v", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": codes.OAuthInvalidData})
		return
	}

	// Social exist, login user
	if social, err := oc.DB.SocialRepository.FindOne(claim.Data.Provider, claim.Data.Id); err == nil {
		token := oc.DB.UserRepository.GenerateTokenFor(social.User)

		// TODO: update social info here

		c.JSON(http.StatusOK, gin.H{"token": token.Token})
		return
	}

	// Procceed with registration
	req := &request.OAuthRegisterRequest{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	// Check if there's no account with this email
	if _, err := oc.DB.UserRepository.GetByEmail(claim.Data.Email); err == nil {
		// TODO: check it
		c.JSON(http.StatusConflict, gin.H{"error": codes.UserExistWithEmail})
		return
	}

	// Validate data
	if errors, err := req.Valid(); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":          err.Error(),
			"errors":         errors,
			"needs_register": true,
		})
		return
	}

	// Check if there's no account with this username
	if _, err := oc.DB.UserRepository.GetByUsername(req.Username); err == nil {
		// TODO: check it
		c.JSON(
			http.StatusConflict,
			gin.H{
				"error": codes.UserExistWithUsername,
				"errors": map[string]string{
					"username": codes.UserExistWithUsername,
				},
			})
		return
	}

	password, err := passwords.HashPassword(req.Password)

	user := &models.User{
		Fullname:    claim.Data.Fetched.Name,
		Username:    req.Username,
		Password:    password,
		Email:       claim.Data.Email,
		Role:        models.USER_ROLES.USER,
		IsActivated: "1",
	}

	oc.DB.UserRepository.Create(user)

	if url := claim.Data.Fetched.Photo; url != "" {
		// TODO: check it
		if photo, err := oc.fileController.UploadRemotePic(url, models.FileTargetProfiles, constants.FileTypeImage, user); err == nil {
			user.Photo = photo
			oc.DB.UserRepository.Save(user)
		}
	}

	social := &models.Social{
		Provider:     claim.Data.Provider,
		AccountId:    claim.Data.Id,
		AccountPhoto: claim.Data.Fetched.Photo,
		AccountName:  claim.Data.Fetched.Name,
		User:         user,
	}

	oc.DB.SocialRepository.Create(social)
	token := oc.DB.UserRepository.GenerateTokenFor(social.User)

	// Send user a token to login
	c.JSON(http.StatusOK, gin.H{"token": token.Token})
}

// List returns users social accounts
func (oc OAuthController) List(c *gin.Context) {
	uid := c.MustGet("UID").(uint)
	list, err := oc.DB.SocialRepository.OfUser(uid)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"accounts": list})
}

func (oc OAuthController) Delete(c *gin.Context) {
	uid := c.MustGet("UID").(uint)
	id := c.Param("id")
	provider := c.Param("provider")

	err := oc.DB.SocialRepository.DeleteOfUser(uid, provider, id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
