package main

import (
	"eba-study/utils"
	"html/template"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ログ出力先
const logfile string = "log/golang.log"

// ログ出力設定の読み込み
func init() {
	utils.LoggingSettings(logfile)
}

func main() {
	e := echo.New()
	initRouting(e)
	e.Logger.Fatal(e.Start(":1323"))
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
return t.templates.ExecuteTemplate(w, name, data)
}

func initRouting(e *echo.Echo) {
	initTemplate(e)
	e.GET("/", index)
}

func initTemplate(e *echo.Echo) {
	t := &Template{
			templates: template.Must(template.ParseGlob("views/*.html")),
	}
	e.Renderer = t
}

func index(c echo.Context) error {
	return c.Render(http.StatusOK, "index", "お勉強同好会")
}
