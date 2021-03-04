package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/muerwre/vault-golang/internal/app"
	"github.com/muerwre/vault-golang/internal/db"
	"github.com/muerwre/vault-golang/internal/db/models"
	"github.com/muerwre/vault-golang/internal/feature/oauth/constants"
	request2 "github.com/muerwre/vault-golang/internal/feature/oauth/request"
	"github.com/muerwre/vault-golang/internal/feature/oauth/usecase"
	utils2 "github.com/muerwre/vault-golang/internal/feature/oauth/utils"
	constants2 "github.com/muerwre/vault-golang/internal/feature/upload/constants"
	usecase2 "github.com/muerwre/vault-golang/internal/feature/upload/usecase"
	userUsecase "github.com/muerwre/vault-golang/internal/feature/user/usecase"
	"github.com/muerwre/vault-golang/internal/service/jwt"
	"github.com/muerwre/vault-golang/pkg/codes"
	"github.com/muerwre/vault-golang/pkg/passwords"
	"github.com/sirupsen/logrus"
	"net/http"
)

type OAuthController struct {
	oauth usecase.OauthUsecase
	user  userUsecase.UserUsecase
	file  usecase2.FileUseCase
}

// TODO: reply to errors via c.HTML in endpoints, which opened in modals

func (oc *OAuthController) Init(db db.DB, config app.Config) *OAuthController {
	oc.oauth = *new(usecase.OauthUsecase).Init(db, config)
	oc.user = *new(userUsecase.UserUsecase).Init(db)
	oc.file = *new(usecase2.FileUseCase).Init(db, config)
	return oc
}

// ProviderMiddleware generates Provider context by :provider url param
func (oc OAuthController) ProviderMiddleware(c *gin.Context) {
	provider := c.Param("provider")
	current, err := utils2.OAuthConfigs.GetByName(provider)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.Set("Provider", current)
}

// Redirect redirects user to oauth endpoint, that redirects back to /process/:target?code= urls
func (oc OAuthController) Redirect(c *gin.Context) {
	provider := c.MustGet("Provider").(*utils2.OAuthConfig)
	url := oc.oauth.GetRedirectUrlForProvider(provider)
	c.Redirect(http.StatusTemporaryRedirect, url)
	return
}

// GetRedirectData is a middleware, that fetches oauth data from provider and passes it further
func (oc OAuthController) GetRedirectData(target string) gin.HandlerFunc {
	return func(c *gin.Context) {
		provider := c.MustGet("Provider").(*utils2.OAuthConfig)
		code := c.Query("code")
		if code == "" {
			utils2.ReplyHtmlEventWithError(c, constants.EventTypeError, codes.OAuthCodeIsEmpty)
			c.Abort()
			return
		}

		data, err := oc.oauth.GetTokenData(provider, code)
		if err != nil {
			logrus.Warnf("Failed to get oauth data: %v", err.Error())
			utils2.ReplyHtmlEventWithError(c, constants.EventTypeError, codes.OAuthInvalidData)
			c.Abort()
			return
		}

		c.Set("UserData", data)
		c.Next()
	}
}

// Process gets fetched from oauth data and encodes it as JWT to send back to frontend, so it can call AttachConfirm with it
func (oc OAuthController) Process(c *gin.Context) {
	ud := c.MustGet("UserData").(*utils2.OauthUserData)
	claim := new(utils2.OauthUserDataClaim).Init(*ud)
	token, err := jwt.EncodeJwtToken(claim)

	if err != nil {
		logrus.Warnf("Failed to create attach token: %v", err.Error())
		utils2.ReplyHtmlEventWithError(c, constants.EventTypeError, codes.OAuthInvalidData)
		return
	}

	utils2.ReplytHtmlEventWithToken(c, constants.EventTypeProcessed, token)
}

// AttachConfirm gets user oauth data from token and creates social connection for it
func (oc OAuthController) AttachConfirm(c *gin.Context) {
	u := c.MustGet("User").(*models.User)
	claim, err := utils2.DecodeOauthClaimFromRequest(c)

	if err != nil {
		logrus.Warnf("Failed to perform attach confirm: %v", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": codes.OAuthInvalidData})
		return
	}

	if exist, err := oc.oauth.GetSocialById(claim.Data.Provider, claim.Data.Id); err == nil {
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
	if user, err := oc.user.GetByEmail(claim.Data.Email); err == nil && user.ID != u.ID {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": codes.UserExistWithEmail})
		return
	}

	social, err := oc.oauth.CreateSocialFromClaim(*claim, u)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": codes.CantSaveUser})
		return
	}

	c.AbortWithStatusJSON(http.StatusOK, gin.H{"account": social})
}

func (oc OAuthController) Login(c *gin.Context) {
	claim, err := utils2.DecodeOauthClaimFromRequest(c)
	if err != nil {
		logrus.Warnf("Failed to perform login: %v", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": codes.OAuthInvalidData})
		return
	}

	// Social exist, login user
	if social, err := oc.oauth.GetSocialById(claim.Data.Provider, claim.Data.Id); err == nil {
		token, _ := oc.user.GenerateTokenFor(social.User)
		// TODO: update social info here
		c.JSON(http.StatusOK, gin.H{"token": token.Token})
		return
	}

	// Procceed with registration
	req := &request2.OAuthRegisterRequest{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": codes.IncorrectData})
		return
	}

	// Check if there's no account with this email
	if _, err := oc.user.GetByEmail(claim.Data.Email); err == nil {
		// TODO: check it
		c.JSON(http.StatusConflict, gin.H{"error": codes.UserExistWithEmail})
		return
	}

	// ValidatePatchRequest data
	if errors, err := req.Valid(); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error":          err.Error(),
			"errors":         errors,
			"needs_register": true,
		})
		return
	}

	// Check if there's no account with this username
	if _, err := oc.user.GetByUsername(req.Username); err == nil {
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
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &models.User{
		Fullname:    claim.Data.Fetched.Name,
		Username:    req.Username,
		Password:    password,
		Email:       claim.Data.Email,
		Role:        models.USER_ROLES.USER,
		IsActivated: "1",
	}

	oc.user.CreateUser(user)
	social, err := oc.oauth.CreateSocialFromClaim(*claim, user)
	if err != nil {
		logrus.Warnf("Can't create social record:\nclaim: %+v\nuser:%+v\n%s", claim, user, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.CantSaveUser})
		return
	}

	token, err := oc.user.GenerateTokenFor(social.User)
	if err != nil {
		logrus.Warnf("Can't create token record: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.CantSaveUser})
		return
	}

	// Send user a token to login
	c.JSON(http.StatusOK, gin.H{"token": token.Token})

	if url := claim.Data.Fetched.Photo; url != "" {
		// TODO: check it
		if photo, err := oc.file.UploadRemotePic(url, models.FileTargetProfiles, constants2.FileTypeImage, user); err == nil {
			oc.user.UpdateUserPhoto(user, photo)
		}
	}
}

// List returns users social accounts
func (oc OAuthController) List(c *gin.Context) {
	u := c.MustGet("User").(*models.User)
	list, err := oc.oauth.GetSocialsOfUser(u)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"accounts": list})
}

func (oc OAuthController) Delete(c *gin.Context) {
	u := c.MustGet("User").(*models.User)
	id := c.Param("id")
	provider := c.Param("provider")

	if err := oc.oauth.DeleteSocialByUserProviderAndId(u, provider, id); err != nil {
		logrus.Warnf("Can't delete social record for user:\nuser: %+v\nprovider: %s\nid: %s", u, provider, id)
		c.JSON(http.StatusBadRequest, gin.H{"error": codes.CantSaveUser})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
