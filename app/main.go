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
var CacheRecord SiteData

func init() {
	BaseUrl, _ = os.LookupEnv("GOLANG_BACKEND_BASE_URL")
}

func main() {
	/*
		@TODO
		何かおかしい.
		リクエストのたびにmainが実行されて変数の初期化がされる.
	*/
	e := echo.New()

	e.Static("/static", "assets")

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

	initRouting(e)
	e.Logger.Fatal(e.Start(":1323"))
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type SiteContent struct {
	Title       string
	PageContent PageContent
}

type PageContent struct {
	Accounts []Account
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

type SiteResponse struct {
	SiteData SiteData `json:"data"`
}

type SiteData struct {
	ExpireDatetime string    `json:"expire_datetime"`
	Accounts       []Account `json:"accounts"`
}

type Account struct {
	AccountId    int64  `json:"id"`
	RegisterName string `json:"registerName"`
	DisplayName  string `json:"displayName"`
	Class        string `json:"class"`
}

func index(c echo.Context) error {

	(c.Logger()).Info(CacheRecord)

	url := BaseUrl + "/api/site"

	/*
		@TODO
		(c.Echo()).GET()

		いったん標準ライブラリで実装したがecho側で代替の機能が準備されている.
	*/
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

	var data SiteResponse

	if err := json.Unmarshal(byteArray, &data); err != nil {
		logger := c.Logger()
		mes := fmt.Sprintf("Error: %s", err)
		logger.Error(mes)

		return err
	}

	(c.Logger()).Info(data.SiteData)

	CacheRecord = data.SiteData

	content := SiteContent{
		Title: "お勉強同好会メンバー一覧",
		PageContent: PageContent{
			Accounts: CacheRecord.Accounts,
		},
	}

	return c.Render(http.StatusOK, "index", content)
}
