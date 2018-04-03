package get

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"ProxyCacheServer/config"
	resp "ProxyCacheServer/entrypoints/v1/response"

	"github.com/codegangsta/martini-contrib/web"
	"github.com/martini-contrib/encoder"
	memCache "github.com/patrickmn/go-cache"
)

// Обработчик запроса на get
func Get(configuration *config.Configuration, ctx *web.Context, enc encoder.Encoder, mc *memCache.Cache, w http.ResponseWriter) (int, []byte) {
	// Подготовка структуры Request из GET параметров
	request := PrepareRequest(configuration, ctx)
	if _, err := request.Validate(); err != nil {
		return resp.Fault(enc, err)
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
			return resp.Fault(enc, err)
		}

		// Обработчик callback на уборщик мусора из кеша (уборщик устаревшего кеша)
		mc.OnEvicted(func(key string, data interface{}) {
			// Получаем структуру кеша
			cache := data.(*Cache)

			// Формирует кеш ответа на запрос
			_, err := MakeResponseCache(key, cache.Request, mc)
			if err != nil {
				panic(err.Error())
			}
		})
	}

	// Пробрасываем список HTTP-заголовков в ответ
	for name, _ := range response.Response.Header {
		w.Header().Set(name, response.Response.Header.Get(name))
	}

	// Пробрасываем нативный ответ от запрашиваемого URL
	return response.Response.StatusCode, []byte(response.Body)
}

// Подготовка структуры Request из GET параметров
func PrepareRequest(configuration *config.Configuration, ctx *web.Context) *Request {
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
func MakeRequest(request *Request) ([]byte, *http.Response, error) {
	// Структура HTTP-клиента
	client := http.Client{}

	// Формируем запрос
	req, err := http.NewRequest("GET", request.Url, nil)
	if err != nil {
		return nil, nil, err
	}

	// Готовим HTTP-данные для запроса
	req.Method = request.Request.Method
	req.Proto = request.Request.Proto
	req.ProtoMajor = request.Request.ProtoMajor
	req.ProtoMinor = request.Request.ProtoMinor

	// Принудительно назначаем заголовок: Accept-Encoding = deflate
	// чтобы небыло геммора с gzip содержимым
	request.Request.Header.Set("Accept-Encoding", "deflate")

	// Пробрасываем список HTTP-заголовков на опрашиваемый сервер
	for name, _ := range request.Request.Header {
		req.Header.Set(name, request.Request.Header.Get(name))
	}

	// Совершаем запрос по указанному URL
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	// После закроем соединение
	// by http://grokbase.com/t/gg/golang-ru/161exdg2tm/http-accept-error-accept-tcp-accept4-too-many-open-files-retrying-in-5ms
	defer resp.Body.Close()

	// Получаем контент запрашиваемого хоста
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	return body, resp, nil
}

// Формирует кеш ответа на запрос
func MakeResponseCache(cacheId string, request *Request, mc *memCache.Cache) (*Response, error) {
	// Делает запрос по указанному URL
	body, resp, err := MakeRequest(request)
	if err != nil {
		return nil, err
	}

	// Структура ответа
	response := NewResponse(string(body), resp)
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
