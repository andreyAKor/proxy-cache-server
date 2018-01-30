package get

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"ProxyCacheServer/config"

	"github.com/codegangsta/martini-contrib/web"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/encoder"
	memCache "github.com/patrickmn/go-cache"
)

// Обработчик запроса на get
func Get(configuration *config.Configuration, ctx *web.Context, enc encoder.Encoder, mc *memCache.Cache, params martini.Params) (int, []byte) {
	// Подготовка структуры Request из GET параметров
	request := PrepareRequest(configuration, ctx)
	if _, err := request.Validate(); err != nil {
		return SendError(enc, err)
	}

	// Идентификатор кеша
	cacheId := request.Url

	// Ответ
	response := &Response{}

	// Получаем ответ из кеша.
	if data, found := mc.Get(cacheId); found {
		// Получаем структуру кеша
		cache := data.(*Cache)

		// Получаем структуру ответа
		response = cache.Response
	} else { // Если в кеше ответов нету
		var err error

		// Формирует кеш ответа на запрос
		response, err = MakeResponseCache(cacheId, request, mc)
		if err != nil {
			return SendError(enc, err)
		}

		// Обработчик callback на уборщик мусора из кеша (уборщик устаревшего кеша)
		mc.OnEvicted(func(key string, data interface{}) {
			fmt.Printf("OnEvicted: %v\n", key)

			// Получаем структуру кеша
			cache := data.(*Cache)

			// Формирует кеш ответа на запрос
			_, err := MakeResponseCache(key, cache.Request, mc)
			if err != nil {
				panic(err.Error())
			}
		})
	}

	return http.StatusOK, encoder.Must(enc.Encode(response))
}

// Формирует контент об ошибке
func SendError(enc encoder.Encoder, err error) (int, []byte) {
	return http.StatusBadRequest, encoder.Must(enc.Encode(map[string]string{
		"error": err.Error(),
	}))
}

// Подготовка структуры Request из GET параметров
func PrepareRequest(configuration *config.Configuration, ctx *web.Context) *Request {
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

	// Формируем структуру запроса
	request := NewRequest(params["url"], configuration.Request.Interval, ctx.Request)

	// Интервал (периодичность) опроса URL-адреса в секундах
	if len(ctx.Params["inteval"]) > 0 {
		request.Inteval, _ = strconv.Atoi(ctx.Params["inteval"])
	}

	return request
}

// Делает запрос по указанному URL
func MakeRequest(request *Request) ([]byte, error) {
	/*
		// by http://polyglot.ninja/golang-making-http-requests/
		resp, err := http.Get(request.Url)
		if err != nil {
			return nil, err
		}
	*/

	// Структура HTTP-клиента
	client := http.Client{}

	// Формируем запрос
	req, err := http.NewRequest("GET", request.Url, nil)
	if err != nil {
		return nil, err
	}

	// Готовим HTTP-данные для запроса
	req.Method = request.Request.Method
	req.Proto = request.Request.Proto
	req.ProtoMajor = request.Request.ProtoMajor
	req.ProtoMinor = request.Request.ProtoMinor

	// Принудительно назначаем заголовок: Accept-Encoding = deflate
	// чтобы небыло геммора с gzip содержимым
	request.Request.Header.Set("Accept-Encoding", "deflate")

	// Пробрасываем список HTTP-заголовков
	for name, _ := range request.Request.Header {
		req.Header.Set(name, request.Request.Header.Get(name))
	}

	// Совершаем запрос по указанному URL
	resp, err := client.Do(req)
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

// Формирует кеш ответа на запрос
func MakeResponseCache(cacheId string, request *Request, mc *memCache.Cache) (*Response, error) {
	// Делает запрос по указанному URL
	value, err := MakeRequest(request)
	if err != nil {
		return nil, err
	}

	// Структура ответа
	response := NewResponse(string(value))
	if _, err := response.Validate(); err != nil {
		return nil, err
	}

	// Данные кеша
	cache := NewCache(request, response)
	if _, err := cache.Validate(); err != nil {
		return nil, err
	}

	// Храним значение переменной в кэше
	mc.Set(cacheId, cache, (time.Duration(request.Inteval) * time.Second))

	return response, nil
}
