package main

import (
	"fmt"
	"github.com/google/uuid"
	yahoo2 "github.com/thethan/fdr-users/internal/yahoo"
	"golang.org/x/oauth2"
	"net/http"
	"os"
)

var (
	oauthConfig *oauth2.Config
)

func init() {
	oauthConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("YAHOO_CLIENT_REDIRECT"),
		ClientID:     os.Getenv("YAHOO_CLIENT_ID"),
		ClientSecret: os.Getenv("YAHOO_CLIENT_SECRET"),
		Scopes:       []string{"fspt-w"},
		Endpoint:     yahoo2.Endpoint,
	}
}

var (
	oauthStateString = uuid.New().String()
)

func main() {
	http.HandleFunc("/", handleMain)
	http.HandleFunc("/login", HandleYahooLogin)
	http.ListenAndServe(":8080", nil)
}

func handleMain(w http.ResponseWriter, r *http.Request) {
	var htmlIndex = `<html>
<body>
	<a href="/login">Google Log In</a>
</body>
</html>`
	fmt.Fprintf(w, htmlIndex)
}

func HandleYahooLogin(w http.ResponseWriter, r *http.Request) {
	url := oauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}