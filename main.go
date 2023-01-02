package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/shenshouer/cas"
	"github.com/valyala/fasthttp/fasthttpadaptor"
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

	app := fiber.New()

	app.Use(adaptor.HTTPMiddleware(client.Handle), requestid.New(), logger.New())

	app.Get("/a", adaptor.HTTPHandlerFunc(handle))
	app.Get("/", handleFiber)

	app.Listen(":9999")

}

func handle(w http.ResponseWriter, r *http.Request) {

	binding := &templateBinding{
		Username:   cas.Username(r),
		Attributes: cas.Attributes(r),
	}
	log.Println(binding)
	log.Println(cas.Username(r))

}

func handleFiber(c *fiber.Ctx) error {
	var r http.Request
	fc := c.Context()

	err := fasthttpadaptor.ConvertRequest(fc, &r, true)
	if err != nil {
		panic(err)
	}

	binding := &templateBinding{
		Username:   cas.Username(&r),
		Attributes: cas.Attributes(&r),
	}

	if !cas.IsAuthenticated(&r) {
		log.Println("is not authenticated!")
		return c.Redirect(loginURL)
	}

	log.Println(binding)
	return c.SendString("username: " + binding.Username)
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
