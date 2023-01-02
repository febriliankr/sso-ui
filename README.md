## BACKUP CODE

package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/go-chi/chi"
	"github.com/shenshouer/cas"
)

var casURL = "https://sso.ui.ac.id/cas2"
var logoutURL = "https://sso.ui.ac.id/cas2/logout"

const appURL = "http://localhost:9999"

var loginURL = fmt.Sprintf("https://sso.ui.ac.id/cas2/login?service=%s&renew=false", appURL)

type templateBinding struct {
	Username   string
	Attributes cas.UserAttributes
}

func main() {
	url, _ := url.Parse(casURL)
	client := cas.NewClient(&cas.Options{
		URL: url,
	})

	root := chi.NewRouter()
	root.Use(client.Handle)

	server := &http.Server{
		Addr:    ":9999",
		Handler: client.Handle(root),
	}

	root.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")

		tmpl, err := template.New("index.html").Parse(index_html)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, error_500, err)
			return
		}

		binding := &templateBinding{
			Username:   cas.Username(r),
			Attributes: cas.Attributes(r),
		}

		if cas.Username(r) == "" {
			log.Println("user is not logged in")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "go to login page", err)
			return
		}

		log.Println(binding)

		html := new(bytes.Buffer)
		if err := tmpl.Execute(html, binding); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, error_500, err)
			return
		}

		html.WriteTo(w)
	})

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

const index_html = `<!DOCTYPE html>
<html>
  <head>
    <title>Welcome {{.Username}}</title>
  </head>
  <body>
    <h1>Welcome {{.Username}} <a href="/logout">Logout</a></h1>
    <p>Your attributes are:</p>
    <ul>{{range $key, $values := .Attributes}}
      <li>{{$len := len $values}}{{$key}}:{{if gt $len 1}}
        <ul>{{range $values}}
          <li>{{.}}</li>{{end}}
        </ul>
      {{else}} {{index $values 0}}{{end}}</li>{{end}}
    </ul>
  </body>
</html>
`

const error_500 = `<!DOCTYPE html>
<html>
  <head>
    <title>Error 500</title>
  </head>
  <body>
    <h1>Error 500</h1>
    <p>%v</p>
  </body>
</html>
`
