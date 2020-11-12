package yahoo

import (
	yahoo3 "github.com/thethan/fdr-users/internal/yahoo"
	"golang.org/x/oauth2"
	"os"
)

func Oauth2Config() *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  os.Getenv("YAHOO_CLIENT_REDIRECT"),
		ClientID:     os.Getenv("YAHOO_CLIENT_ID"),
		ClientSecret: os.Getenv("YAHOO_CLIENT_SECRET"),
		Scopes:       []string{"fspt-w"},
		Endpoint:     yahoo3.Endpoint,
	}
}
