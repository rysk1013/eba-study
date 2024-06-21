package main

import (
	"eba-study/utils"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var BaseUrl string
var CacheRecord SiteData
var DisplayClass DisplayClassData

const api_path string = "/api/site"

type DisplayClassData struct {
	DC01 string `json:"01"`
	DC02 string `json:"02"`
}

func init() {
	BaseUrl, _ = os.LookupEnv("GOLANG_BACKEND_BASE_URL")

	f, err := os.Open("display_class.json")
	if err != nil {
		panic(err)
	}

	defer f.Close()

	b, _ := io.ReadAll(f)

	json.Unmarshal(b, &DisplayClass)
}

func main() {

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

	func() {
		if err := e.Start(":1323"); err != nil {
			e.Logger.Info(err)
			e.Logger.Info("shutting down the server")
		}
	}()
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
	RegisterName string `json:"register_name"`
	DisplayName  string `json:"display_name"`
	Class        string `json:"class"`
	DisplayClass string `json:"display_class"`
}

func (ac Account) GetInlineStyle() string {
	/*
	 * @TODO
	 * 表示の定義はjsonフォーマットで外部ファイルに切り出す.
	 * プロセス起動時にメモリ上に配置して使用する.
	 *
	 * あんま意味ない、あとで戻す
	 */
	switch ac.DisplayClass {
	case "01":
		return DisplayClass.DC01
	case "02":
		return DisplayClass.DC02
	default:
		return ""
	}
}

func index(c echo.Context) error {

	e := c.Echo()

	var t = CacheRecord.ExpireDatetime
	expireTime, _ := time.Parse("20060102150405", t)

	if expireTime.Before(time.Now()) {

		var data SiteResponse

		if err := InvokeApi(&data); err != nil {

			e.Logger.Error(fmt.Sprintf("Error: %s", err))

			return err
		}

		CacheRecord = data.SiteData
	}

	/*
		for _, v := range data.SiteData.Accounts {
			e.Logger.Info(v)
		}
	*/

	content := SiteContent{
		Title: "お勉強同好会メンバー一覧",
		PageContent: PageContent{
			Accounts: CacheRecord.Accounts,
		},
	}

	return c.Render(http.StatusOK, "index", content)
}

func InvokeApi(data *SiteResponse) error {

	url := BaseUrl + api_path

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {

		return err
	}

	client := new(http.Client)

	resp, err := client.Do(req)
	if err != nil {

		return err
	}

	defer resp.Body.Close()

	byteArray, _ := io.ReadAll(resp.Body)

	if err := json.Unmarshal(byteArray, &data); err != nil {

		return err
	}

	return nil
}
