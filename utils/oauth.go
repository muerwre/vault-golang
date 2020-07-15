package utils

import (
	"fmt"
	"github.com/muerwre/vault-golang/utils/codes"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"strconv"
)

const (
	ProviderVk     string = "vk"
	ProviderGoogle string = "google"
)

func GetOauthVkConfig(id string, secret string, redirect string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     id,
		ClientSecret: secret,
		RedirectURL:  redirect,
		Scopes:       []string{"email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://oauth.vk.com/authorize",
			TokenURL: "https://oauth.vk.com/access_token",
		},
	}
}

func GetOauthGoogleConfig(id string, secret string, redirect string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     id,
		ClientSecret: secret,
		RedirectURL:  redirect,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

type OauthUserData struct {
	Provider string
	Id       string
	Email    string
	Token    string
}

func ProcessVkData(token *oauth2.Token) (OauthUserData, error) {
	data := OauthUserData{
		Provider: ProviderVk,
		Id:       strconv.Itoa(int(token.Extra("user_id").(float64))),
		Token:    token.Extra("access_token").(string),
		Email:    token.Extra("email").(string),
	}

	return data, nil
}

func ProcessGoogleData(token *oauth2.Token) (OauthUserData, error) {
	data := OauthUserData{}

	return data, nil
}

type OAuthConfig struct {
	ConfigCreator func(id string, secret string, redirect string) *oauth2.Config
	Parser        func(token *oauth2.Token) (OauthUserData, error)
}

type OAuthConfigList map[string]*OAuthConfig

var OAuthConfigs = OAuthConfigList{
	ProviderVk: &OAuthConfig{
		ConfigCreator: GetOauthVkConfig,
		Parser:        ProcessVkData,
	},
	ProviderGoogle: &OAuthConfig{
		ConfigCreator: GetOauthGoogleConfig,
		Parser:        ProcessGoogleData,
	},
}

func (c OAuthConfigList) GetByName(name string) (*OAuthConfig, error) {
	for k, v := range c {
		if k == name && v != nil {
			return v, nil
		}
	}

	return nil, fmt.Errorf(codes.OAuthUnknownProvider)
}
