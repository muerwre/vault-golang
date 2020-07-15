package controllers

import (
	"context"
	"fmt"
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
}

func (oc OAuthController) ProviderMiddleware(c *gin.Context) {
	provider := c.Param("provider")
	current, err := utils.OAuthConfigs.GetByName(provider)

	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.Set("Provider", current)
}

func (oc OAuthController) Redirect(c *gin.Context) {
	provider := c.MustGet("Provider").(*utils.OAuthConfig)
	config := provider.ConfigCreator(oc.Config.VkClientId, oc.Config.VkClientSecret, oc.Config.VkCallbackUrl)
	c.Redirect(http.StatusTemporaryRedirect, config.AuthCodeURL("pseudo-random")) // TODO: pseudo-random can be in payload
}

func (oc OAuthController) Process(c *gin.Context) {
	ctx := context.Background()
	provider := c.MustGet("Provider").(*utils.OAuthConfig)
	code := c.Query("code")

	if code == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": codes.OAuthCodeIsEmpty})
		return
	}

	config := provider.ConfigCreator(oc.Config.VkClientId, oc.Config.VkClientSecret, oc.Config.VkCallbackUrl)
	token, err := config.Exchange(ctx, code)

	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Failed to get token"})
		logrus.Infof("Failed to get token: %v", err.Error())
		return
	}

	data, err := provider.Parser(token)

	if user, _ := oc.DB.UserRepository.GetByEmail(data.Email); user.ID != 0 {
		c.String(http.StatusOK, fmt.Sprintf("code: %+v", data))
		return
	}

	c.String(http.StatusOK, fmt.Sprintf("code: %+v", data))
	return
}

func (oc OAuthController) ProcessToken(c *gin.Context) {
	ctx := context.Background()
	code := c.Query("code")

	if code == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": codes.OAuthCodeIsEmpty})
		return
	}

	config := utils.GetOauthGoogleConfig(oc.Config.GoogleClientId, oc.Config.GoogleClientSecret, oc.Config.GoogleCallbackUrl)
	token, err := config.Exchange(ctx, code)

	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Failed to get token"})
		logrus.Infof("Failed to get token: %v", err.Error())
		return
	}

	tokenJwt := token.Extra("id_token").(string) // TODO: decode jwt payload, get email, email_verified and name

	c.String(http.StatusOK, fmt.Sprintf("code: %d %s %s", 1, token, tokenJwt))
	return
}

func (oc OAuthController) Attach(c *gin.Context) {
	// TODO: get data by token
	// TODO: if in base (oauth.account.id AND base.user.id != user.id) OR (user with oauth.account.email and base.user.id !+ user.id) -> error
	// TODO: create connection
	c.String(http.StatusOK, "TODO:")
}

func (oc OAuthController) Login(c *gin.Context) {
	// TODO: get data by token
	// TODO: we have such connection? yes: login, exit, sending tokens
	// TODO: we have user with this email? yes: exit, sending user's name and waiting for ATTACH
	// TODO: register new user, create connection
	c.String(http.StatusOK, "TODO:")
}

//url := fmt.Sprintf(
//	`https://api.vk.com/method/users.get?user_id=%s&fields=photo,email&v=5.67&access_token=%s`,
//	fmt.Sprintf("%v", int(token.Extra("user_id").(float64))),
//	token.AccessToken,
//)

//response, err := http.Get(url)

//if err != nil {
//	c.JSON(http.StatusForbidden, gin.H{"error": "Failed getting user info"})
//	logrus.Infof("Failed getting user info: %v", err.Error())
//	return
//}

//defer response.Body.Close()

//contents, err := ioutil.ReadAll(response.Body)

//if err != nil {
//	c.JSON(http.StatusForbidden, gin.H{"error": "Failed to read response"})
//	return
//}

//data := &request.VkApiRequest{}

//err = json.Unmarshal(contents, &data)

//if data.Response == nil || err != nil {
//	c.JSON(http.StatusForbidden, gin.H{"error": "Can't get user"})
//
//	return
//}

//println("response is", data)

// TODO: just give this token back to frontend. Create some endpoint like /oauth/vk/register and /oauth/vk/attach
// TODO: and use token there

//user, err := d.FindOrCreateUser(
//	&model.User{
//		Uid:   fmt.Sprintf("vk:%d", data.Response[0].Id),
//		Name:  fmt.Sprintf("%s %s", data.Response[0].FirstName, data.Response[0].LastName),
//		Photo: fmt.Sprintf("%v", data.Response[0].Photo),
//		Role:  "vk",
//	},
//)
//
//if err != nil {
//	c.JSON(http.StatusForbidden, gin.H{"error": "Can't get user"})
//	return
//}
//
//random_url := d.GenerateRandomUrl()
//
//c.HTML(http.StatusOK, "social.html", AuthResponse{User: user, RandomUrl: random_url})
