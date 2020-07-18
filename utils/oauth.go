package utils

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/muerwre/vault-golang/utils/codes"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	ProviderVk     string = "vk"
	ProviderGoogle string = "google"
)

type OauthUserData struct {
	Provider string
	Id       string
	Email    string
	Token    string
}

type OAuthConfig struct {
	ConfigCreator func(credentials OAuthCredentials) *oauth2.Config
	Parser        func(token *oauth2.Token) (*OauthUserData, error)
	Fetcher       func(token string) (*OAuthFetchResult, error)
}

type OAuthConfigList map[string]*OAuthConfig

var OAuthConfigs = OAuthConfigList{
	ProviderVk: &OAuthConfig{
		ConfigCreator: GetOauthVkConfig,
		Parser:        ProcessVkData,
		Fetcher:       FetchVkData,
	},
	ProviderGoogle: &OAuthConfig{
		ConfigCreator: GetOauthGoogleConfig,
		Parser:        ProcessGoogleData,
		Fetcher:       FetchGoogleData,
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

type VkResponse struct {
	Response []struct {
		Id        int    `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Photo     string `json:"photo"`
	} `json:"response"`
}

type OAuthFetchResult struct {
	Provider string
	Id       int
	Name     string
	Photo    string
}

func GetOauthVkConfig(credentials OAuthCredentials) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     credentials.VkClientId,
		ClientSecret: credentials.VkClientSecret,
		RedirectURL:  credentials.VkCallbackUrl,
		Scopes:       []string{"email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://oauth.vk.com/authorize",
			TokenURL: "https://oauth.vk.com/access_token",
		},
	}
}

func GetOauthGoogleConfig(credentials OAuthCredentials) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     credentials.GoogleClientId,
		ClientSecret: credentials.GoogleClientSecret,
		RedirectURL:  credentials.GoogleCallbackUrl,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

func ProcessVkData(token *oauth2.Token) (*OauthUserData, error) {
	data := &OauthUserData{
		Provider: ProviderVk,
		Id:       strconv.Itoa(int(token.Extra("user_id").(float64))),
		Token:    token.Extra("access_token").(string),
		Email:    token.Extra("email").(string),
	}

	return data, nil
}

func ProcessGoogleData(token *oauth2.Token) (*OauthUserData, error) {
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

func (c OAuthConfigList) GetByName(name string) (*OAuthConfig, error) {
	for k, v := range c {
		if k == name && v != nil {
			return v, nil
		}
	}

	return nil, fmt.Errorf(codes.OAuthUnknownProvider)
}

func FetchVkData(code string) (*OAuthFetchResult, error) {
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

	data := &VkResponse{}
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

func FetchGoogleData(code string) (*OAuthFetchResult, error) {
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + code)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	println(contents)

	return nil, nil
}
