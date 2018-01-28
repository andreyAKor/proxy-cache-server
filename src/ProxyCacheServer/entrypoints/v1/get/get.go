package get

import (
	//"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	//"ProxyCacheServer/structs"
	"ProxyCacheServer/config"

	"github.com/codegangsta/martini-contrib/web"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/encoder"
	memCache "github.com/patrickmn/go-cache"
)

// Обработчик запроса на get
func Get(configuration *config.Configuration, ctx *web.Context, enc encoder.Encoder, cache *memCache.Cache, params martini.Params) (int, []byte) {
	// Структура запроса
	request := PrepareRequestFromWebContext(ctx)
	if _, err := request.Validate(); err != nil {
		return http.StatusBadRequest, encoder.Must(enc.Encode(map[string]string{
			"error": err.Error(),
		}))
	}

	// TODO
	value, err := MakeRequest(request)
	if err != nil {
		return http.StatusBadRequest, encoder.Must(enc.Encode(map[string]string{
			"error": err.Error(),
		}))
	}

	// Структура ответа
	response := NewResponse(string(value))
	if _, err := response.Validate(); err != nil {
		return http.StatusBadRequest, encoder.Must(enc.Encode(map[string]string{
			"error": err.Error(),
		}))
	}

	return http.StatusOK, encoder.Must(enc.Encode(response))
}

// Подготовка структуры Request из GET параметров
func PrepareRequestFromWebContext(ctx *web.Context) *Request {
	/*
		fmt.Printf("======================================================\n")
		fmt.Printf("UserAgent: %v\n", ctx.Request.UserAgent())
		fmt.Printf("Method: %v\n", ctx.Request.Method)
		fmt.Printf("Proto: %v\n", ctx.Request.Proto)
		fmt.Printf("Referer: %v\n", ctx.Request.Referer())
		//fmt.Printf("BasicAuth: %v\n", ctx.Request.BasicAuth())
		fmt.Printf("Cookies: %v\n", ctx.Request.Cookies())
		fmt.Printf("Header: %v\n", ctx.Request.Header)
		fmt.Printf("======================================================\n")
	*/

	// Обязательные GET-параметры
	params := map[string]string{
		"url": "",
	}

	// Опрашиваемый URL-адрес
	if len(ctx.Params["url"]) > 0 {
		params["url"], _ = url.QueryUnescape(ctx.Params["url"])
	}

	request := NewRequest(params["url"], 10, ctx.Request.UserAgent(), ctx.Request.Method, ctx.Request.Proto)

	// Интервал (периодичность) опроса URL-адреса в секундах
	if len(ctx.Params["inteval"]) > 0 {
		request.Inteval, _ = strconv.Atoi(ctx.Params["inteval"])
	}

	// HTTP содержимое клиента
	request.Referer = ctx.Request.Referer()
	request.BasicAuthUsername, request.BasicAuthPassword, _ = ctx.Request.BasicAuth()
	request.Cookies = ctx.Request.Cookies()
	request.Header = ctx.Request.Header

	return request
}

// Делает запрос по указанному URL внешнего API
func MakeRequest(request *Request) ([]byte, error) {
	// by http://polyglot.ninja/golang-making-http-requests/
	resp, err := http.Get(request.Url)
	if err != nil {
		return nil, err
	}

	// После закроем соединение
	// by http://grokbase.com/t/gg/golang-ru/161exdg2tm/http-accept-error-accept-tcp-accept4-too-many-open-files-retrying-in-5ms
	defer resp.Body.Close()

	// Получаем контент запрашиваемого хоста
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
