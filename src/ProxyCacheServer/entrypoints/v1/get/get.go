package get

import (
	"net/http"
	"net/url"
	"strconv"

	//"ProxyCacheServer/structs"

	"github.com/codegangsta/martini-contrib/web"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/encoder"
	memCache "github.com/patrickmn/go-cache"
)

// Обработчик запроса на поиск
func Get(ctx *web.Context, enc encoder.Encoder, cache *memCache.Cache, params martini.Params) (int, []byte) {
	// Структура запроса
	request := PrepareRequestFromGetParams(ctx.Params)

	// Если во время валидации была ошибка, то формируем панику на основе первой ошибки
	if _, err := request.Validate(); err != nil {
		panic(err.Error())
	}

	// TODO

	// Структура ответа
	response := NewResponse(request.Url)

	// Если во время валидации была ошибка, то формируем панику на основе первой ошибки
	if _, err := response.Validate(); err != nil {
		panic(err.Error())
	}

	return http.StatusOK, encoder.Must(enc.Encode(response))
}

// Подготовка структуры Request из GET параметров
func PrepareRequestFromGetParams(params map[string]string) *Request {
	// Декодируем строки
	params["url"], _ = url.QueryUnescape(params["url"])

	request := NewRequest(params["url"])

	if len(params["limit"]) > 0 {
		request.Limit, _ = strconv.Atoi(params["limit"])
	}

	return request
}
