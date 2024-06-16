package main

import (
	"eba-study/utils"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var BaseUrl string
var SiteData site_data

func init() {
	BaseUrl, _ = os.LookupEnv("GOLANG_BACKEND_BASE_URL")
}

func main() {
	e := echo.New()

	logfile, ok := os.LookupEnv("GOLANG_LOG_FILE")
	if !ok {
		panic("GOLANG_LOG_FILE is not set")
	}

	fp, err := utils.GetFilePointer(logfile)
	if err != nil {
		panic(err)
	}

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "time=${time_rfc3339_nano}, method=${method}, uri=${uri}, status=${status}\n",
		Output: fp,
	}))

	e.Logger.SetOutput(fp)

	loglevel, _ := os.LookupEnv("GOLANG_LOG_LEVEL")
	e.Logger.SetLevel(utils.GetLEVEL(loglevel))

	e.Logger.Info("## Begin ##")

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

type site_response struct {
	SiteData site_data `json:"data"`
}

type site_data struct {
	ExpireDatetime string    `json:"expire_datetime"`
	Accounts       []account `json:"accounts"`
}

type account struct {
	AccountId    int64  `json:"id"`
	RegisterName string `json:"registerName"`
	DisplayName  string `json:"displayName"`
	Class        string `json:"class"`
}

func index(c echo.Context) error {

	(c.Logger()).Info(SiteData)

	url := BaseUrl + "/api/site"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger := c.Logger()
		mes := fmt.Sprintf("Error: %s", err)
		logger.Error(mes)

		return err
	}

	client := new(http.Client)

	resp, err := client.Do(req)
	if err != nil {
		logger := c.Logger()
		mes := fmt.Sprintf("Error: %s", err)
		logger.Error(mes)

		return err
	}

	defer resp.Body.Close()

	byteArray, _ := io.ReadAll(resp.Body)

	var data site_response

	if err := json.Unmarshal(byteArray, &data); err != nil {
		logger := c.Logger()
		mes := fmt.Sprintf("Error: %s", err)
		logger.Error(mes)

		return err
	}

	(c.Logger()).Info(data.SiteData)

	SiteData = data.SiteData

	return c.Render(http.StatusOK, "index", "お勉強同好会")
}
