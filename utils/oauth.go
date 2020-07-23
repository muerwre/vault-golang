package utils

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/muerwre/vault-golang/request"
	"github.com/muerwre/vault-golang/utils/codes"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

const (
	ProviderVk          string = "vkontakte"
	ProviderGoogle      string = "google"
	ProcessTargetAttach string = "attach"
	ProcessTargetLogin  string = "login"
	ClaimPayloadType    string = "oauth_claim"
)

type OauthUserData struct {
	Provider string
	Id       string
	Email    string
	Token    string
	Fetched  *OAuthFetchResult
}

type OauthUserDataClaim struct {
	Data OauthUserData
	Type string
}

func (d OauthUserDataClaim) Valid() error {
	if d.Type != ClaimPayloadType {
		return fmt.Errorf("Invalid claim type.")
	}

	return nil
}

func (d *OauthUserDataClaim) Init(data OauthUserData) *OauthUserDataClaim {
	d.Type = ClaimPayloadType
	d.Data = data
	return d
}

type OAuthConfig struct {
	ConfigCreator func(credentials OAuthCredentials, target string) *oauth2.Config
	Parser        func(token *oauth2.Token) (*OauthUserData, error)
	Fetcher       func(token string) (*OAuthFetchResult, error)
}

type oAuthConfigList map[string]*OAuthConfig

var OAuthConfigs = oAuthConfigList{
	ProviderVk: &OAuthConfig{
		ConfigCreator: getOauthVkConfig,
		Parser:        processVkData,
		Fetcher:       fetchVkData,
	},
	ProviderGoogle: &OAuthConfig{
		ConfigCreator: getOauthGoogleConfig,
		Parser:        processGoogleData,
		Fetcher:       fetchGoogleData,
	},
}

type OAuthCredentials struct {
	VkClientId         string
	VkClientSecret     string
	VkCallbackUrl      string
	GoogleClientId     string
	GoogleClientSecret string
	GoogleCallbackUrl  string
}

type vkResponse struct {
	Response []struct {
		Id        int    `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Photo     string `json:"photo"`
	} `json:"response"`
}

type googleResponse struct {
	Id      string
	Email   string
	Name    string
	Picture string
}

type OAuthFetchResult struct {
	Provider string
	Id       int
	Name     string
	Photo    string
}

func getOauthVkConfig(credentials OAuthCredentials, target string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     credentials.VkClientId,
		ClientSecret: credentials.VkClientSecret,
		RedirectURL:  strings.Join([]string{credentials.VkCallbackUrl, target}, "/"),
		Scopes:       []string{"email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://oauth.vk.com/authorize",
			TokenURL: "https://oauth.vk.com/access_token",
		},
	}
}

func getOauthGoogleConfig(credentials OAuthCredentials, target string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     credentials.GoogleClientId,
		ClientSecret: credentials.GoogleClientSecret,
		RedirectURL:  strings.Join([]string{credentials.GoogleCallbackUrl, target}, "/"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

func processVkData(token *oauth2.Token) (*OauthUserData, error) {
	data := &OauthUserData{
		Provider: ProviderVk,
		Id:       strconv.Itoa(int(token.Extra("user_id").(float64))),
		Token:    token.Extra("access_token").(string),
		Email:    token.Extra("email").(string),
	}

	return data, nil
}

func processGoogleData(token *oauth2.Token) (*OauthUserData, error) {
	tokenStr := token.Extra("id_token").(string)

	tok, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})

	if err != nil {
		return nil, err
	}

	claims, ok := tok.Claims.(jwt.MapClaims)

	if !ok || !token.Valid() {
		return nil, fmt.Errorf(codes.OAuthInvalidData)
	}

	data := &OauthUserData{
		Provider: ProviderGoogle,
		Token:    token.AccessToken,
		Email:    claims["email"].(string),
		Id:       claims["email"].(string),
	}

	return data, nil
}

func (c oAuthConfigList) GetByName(name string) (*OAuthConfig, error) {
	for k, v := range c {
		if k == name && v != nil {
			return v, nil
		}
	}

	return nil, fmt.Errorf(codes.OAuthUnknownProvider)
}

func fetchVkData(code string) (*OAuthFetchResult, error) {
	url := fmt.Sprintf(
		`https://api.vk.com/method/users.get?user_id=%s&fields=photo,email&v=5.67&access_token=%s`,
		"360004",
		code,
	)

	response, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	data := &vkResponse{}
	err = json.Unmarshal(contents, &data)

	if err != nil {
		return nil, err
	}

	return &OAuthFetchResult{
		Provider: ProviderVk,
		Id:       data.Response[0].Id,
		Photo:    data.Response[0].Photo,
		Name:     fmt.Sprintf("%s %s", data.Response[0].FirstName, data.Response[0].LastName),
	}, nil
}

func fetchGoogleData(code string) (*OAuthFetchResult, error) {
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + code)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	data := &googleResponse{}
	err = json.Unmarshal(contents, &data)

	if err != nil {
		return nil, err
	}

	id, _ := strconv.Atoi(data.Id)

	return &OAuthFetchResult{
		Provider: ProviderGoogle,
		Id:       id,
		Photo:    data.Picture,
		Name:     data.Name,
	}, nil
}

func DecodeOauthClaimFromRequest(c *gin.Context) (*OauthUserDataClaim, error) {
	req := &request.OAuthAttachConfirmRequest{}

	if err := c.BindJSON(&req); err != nil {
		return nil, err
	}

	result, err := DecodeJwtToken(req.Token, &OauthUserDataClaim{})

	if err != nil {
		return nil, err
	}

	return result.(*OauthUserDataClaim), nil
}
