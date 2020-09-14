package yahoo

import (
	"crypto/rand"
	"encoding/base64"
	"golang.org/x/oauth2"

	"net/http"
	"os"
	"time"
)

var Endpoint = oauth2.Endpoint{
	AuthURL:  "https://api.login.yahoo.com/oauth2/request_auth",
	TokenURL: "https://api.login.yahoo.com/oauth2/get_token",
}

// Scopes: OAuth 2.0 scopes provide a way to limit the amount of access that is granted to an access token.
var OauthConfig = &oauth2.Config{
	Endpoint:     Endpoint,
	RedirectURL:  os.Getenv("YAHOO_CLIENT_REDIRECT"),
	ClientID:     os.Getenv("YAHOO_CLIENT_ID"),
	ClientSecret: os.Getenv("YAHOO_CLIENT_SECRET"),
	Scopes:       []string{"fspt-w"},
}


func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(365 * 24 * time.Hour)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}


