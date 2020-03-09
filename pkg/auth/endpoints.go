package auth

import (
	"context"
	"fmt"
	"github.com/markbates/goth"
	"github.com/thethan/fdr-users/pkg/gothic"
	"html/template"
	"net/http"
	"sort"
)

func NewEndpoints(provider goth.Provider) Endpoints {
	return Endpoints{provider: provider}
}

// Endpoints in this package are not standard go-kut packages.
// They align with the mux handler. These are not
type Endpoints struct {
	provider goth.Provider
}

func decodeFuncNothing(ctx context.Context, req *http.Request) (interface{}, error) {
	return nil, nil
}

// EncodeHTTPGenericResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer. Primarily useful in a server.

func EncodeFuncTemplates(ctx context.Context, w http.ResponseWriter, res interface{}) error {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	templateName, _ := ctx.Value("template").(string)
	t, _ := template.New(templateName).Parse(indexTemplate)
	t.Execute(w, res)
	return nil
}

// @todo move this to an svc
func (e Endpoints) Index(ctx context.Context, req interface{}) (interface{}, error) {
	ctx = context.WithValue(ctx, "template", "index")
	goth.UseProviders(e.provider)
	m := make(map[string]string)

	fmt.Println("this is latest")
	m["yahoo"] = "Yahoo"

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	providerIndex := &ProviderIndex{Providers: keys, ProvidersMap: m}

	return providerIndex, nil
}

// @todo move this to an endpoint
func (e Endpoints) CompleteUserAuthHandler(res http.ResponseWriter, req *http.Request) {

	goth.UseProviders(e.provider)

	user, err := gothic.CompleteUserAuth(res, req)
	if err != nil {
		fmt.Println(res, err)
		return
	}
	t, _ := template.New("user").Parse(userTemplate)
	t.Execute(res, user)
}

// @todo move this to an endpoint
func (e Endpoints) Logout(res http.ResponseWriter, req *http.Request) {
	_ = gothic.Logout(res, req)
	res.Header().Set("Location", "/")
	res.WriteHeader(http.StatusTemporaryRedirect)
}

// @todo move this to an endpoint
func (e Endpoints) Provider(res http.ResponseWriter, req *http.Request) {

	//ctx = context.WithValue(ctx, "template", "user")
	if gothUser, err := gothic.CompleteUserAuth(res, req); err == nil {
		fmt.Println("err is nil in getting user")
		t, _ := template.New("user").Parse(userTemplate)
		t.Execute(res, gothUser)
		return
	} else {
		fmt.Println("Beginning Auth")
		fmt.Println(err)
		gothic.BeginAuthHandler(res, req)

	}
}

type ProviderIndex struct {
	Providers    []string
	ProvidersMap map[string]string
}

var indexTemplate = `{{range $key,$value:=.Providers}}
    <p><a href="/users/users/auth/{{$value}}/login">Log in with {{index $.ProvidersMap $value}}</a></p>
{{end}}`

var userTemplate = `
<p><a href="/logout/{{.Provider}}">logout</a></p>
<p>Name: {{.Name}} [{{.LastName}}, {{.FirstName}}]</p>
<p>Email: {{.Email}}</p>
<p>NickName: {{.NickName}}</p>
<p>Location: {{.Location}}</p>
<p>AvatarURL: {{.AvatarURL}} <img src="{{.AvatarURL}}"></p>
<p>Description: {{.Description}}</p>
<p>UserID: {{.UserID}}</p>
<p>AccessToken: {{.AccessToken}}</p>
<p>ExpiresAt: {{.ExpiresAt}}</p>
<p>RefreshToken: {{.RefreshToken}}</p>
`

func (e Endpoints) CompleteUserAuthEndpoint(ctx context.Context, request interface{}) (response interface{}, err error) {
	return nil, nil
}
