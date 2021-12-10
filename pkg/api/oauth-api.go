package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"pdserver/pkg/api/model"
	"time"

	"golang.org/x/oauth2"
)

type NaverOAuth struct {
	Config *oauth2.Config
	State  string
}

type NaverAPIService interface {
	Auth(http.ResponseWriter, *http.Request) (*model.Token, error)
	GetUser(*model.Token) (*model.User, error)
}

// NewNaverService - Create Naver OAuth
func NewNaverService() *NaverOAuth {
	config := &oauth2.Config{
		ClientID:     os.Getenv("NAVER_ID"),
		ClientSecret: os.Getenv("NAVER_SECRET"),
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://nid.naver.com/oauth2.0/authorize",
			TokenURL: "https://nid.naver.com/oauth2.0/token",
		},
		RedirectURL: "http://localhost:8080/auth/naver/callback",
		Scopes:      []string{"email", "name", "nickname", "birth"},
	}
	return &NaverOAuth{
		Config: config,
		State:  os.Getenv("ACCESS_SECRET"),
	}
}

// Auth - Get tokens from naver
func (n *NaverOAuth) Auth(w http.ResponseWriter, r http.Request) (*model.Token, error) {
	if r.FormValue("state") != n.State {
		return nil, fmt.Errorf("Invalid oauth state")
	}
	code := r.FormValue("code")
	client := &http.Client{Timeout: 2 * time.Second}
	ctx := context.WithValue(oauth2.NoContext, oauth2.HTTPClient, client)

	token, err := n.Config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	return &model.Token{AccessToken: token.AccessToken, RefreshToken: token.RefreshToken}, nil
}

// GetUser - Get user from naver
func (n *NaverOAuth) GetUser(token *model.Token) (*model.User, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://openapi.naver.com/v1/nid/me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Host", "openapi.naver.com")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "*/*")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var oauthRes model.OAuthResponse
	if err := json.Unmarshal(body, &oauthRes); err != nil {
		return nil, err
	}
	userRes := oauthRes.Response
	return &model.User{Email: userRes.Email, Name: userRes.Name, Nickname: userRes.Nickname, Birth: userRes.Birthyear + "-" + userRes.Birthday}, nil
}
