package main

import (
	"context"
	"eba-study/utils"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

/*
 * @TODO
 * 処理がまとまりすぎているので後でリファクタ.
 */

var Environment string
var BaseUrl string
var RetryLimitThreshold int
var IsUseDummyAccount bool
var CacheRecordLock bool
var CacheRecord SiteData
var DisplayClass DisplayClassData

const api_path string = "/api/site"

type DisplayClassData struct {
	DC01 string `json:"01"`
	DC02 string `json:"02"`
}

func init() {

	Environment, _ = os.LookupEnv("GOLANG_ENVIRONMENT")
	BaseUrl, _ = os.LookupEnv("GOLANG_BACKEND_BASE_URL")
	account_data_retry_limit_threshold, _ := os.LookupEnv("GOLANG_ACCOUNT_DATA_RETRY_LIMIT_THRESHOLD")

	var err error

	RetryLimitThreshold, err = strconv.Atoi(account_data_retry_limit_threshold)
	if err != nil {
		panic("GOLANG_ACCOUNT_DATA_RETRY_LIMIT_THRESHOLD is not set.")
	}

	is_use_dummy_account, isOk := os.LookupEnv("GOLANG_USE_DUMMY_ACCOUNTS_DATA")

	if !isOk {
		is_use_dummy_account = "0"
	}

	IsUseDummyAccount = is_use_dummy_account == "1"

	CacheRecordLock = false

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
		panic("GOLANG_LOG_FILE is not set.")
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

	if Environment == "dev" {
		/*
		 * watchexecでホットリロードを設定する場合、goroutineでは動かない.
		 */
		func() {
			if err := e.Start(":1323"); err != nil {

				e.Logger.Info(err)
				e.Logger.Info("shutting down the server.")
			}
		}()
	} else {
		/*
		 * 並行実行.
		 */
		go func() {
			if err := e.Start(":1323"); err != nil {

				e.Logger.Info(err)
				e.Logger.Info("shutting down the server.")
			}
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	e.Logger.Info("graceful shutting down the server.")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	if err := e.Shutdown(ctx); err != nil {

		e.Logger.Info(err)
		e.Close()
	}

	time.Sleep(1 * time.Second)
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
	case "D1":
		return "dummy first"
	case "D2":
		return "dummy second"
	case "D3":
		return "dummy third"
	default:
		return ""
	}
}

func index(c echo.Context) error {

	e := c.Echo()

	t := CacheRecord.ExpireDatetime
	expireTime, _ := time.Parse("20060102150405", t)

	if expireTime.Before(time.Now()) {

		var data SiteResponse

		retry := RetryLimitThreshold

		for retry > 0 {

			if IsUseDummyAccount {

				raw, err := os.ReadFile("./dummy_accounts.json")
				if err != nil {

					e.Logger.Error(fmt.Sprintf("Error: %s", err))

					retry--

					continue
				}

				if err := json.Unmarshal(raw, &data); err != nil {

					e.Logger.Error(fmt.Sprintf("Error: %s", err))

					retry--

					continue
				}

				sixHour := 6 * time.Hour
				data.SiteData.ExpireDatetime = time.Now().Add(sixHour).Format("20060102150405")

				CacheRecord = data.SiteData

				break
			}

			if CacheRecordLock {

				retry--

				if retry != 0 {

					time.Sleep(1 * time.Second)
				}

				continue
			}

			CacheRecordLock = true

			if err := InvokeApi(&data); err != nil {

				CacheRecordLock = false

				e.Logger.Error(fmt.Sprintf("Error: %s", err))

				return err
			}

			CacheRecord = data.SiteData

			CacheRecordLock = false
		}

		t := CacheRecord.ExpireDatetime
		expireTime, _ := time.Parse("20060102150405", t)

		if expireTime.Before(time.Now()) {
			/**
			 * @TODO
			 * データ更新失敗時に1日1回アラートメールを送信してもいい.
			 *
			 * @TODO
			 * 更新に失敗した場合に古いデータが残っているなら古いデータで処理を継続することを検討してもいい.
			 */
			err := errors.New(fmt.Sprintf("Error: %s", "CacheRecord update failed."))

			e.Logger.Error(err)

			return err
		}
	}

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
