package get

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/andreyAKor/proxy-cache-server/internal/config"
	resp "github.com/andreyAKor/proxy-cache-server/internal/entrypoints/v1/response"

	"github.com/codegangsta/martini-contrib/web"
	"github.com/martini-contrib/encoder"
	memCache "github.com/patrickmn/go-cache"
)

// Обработчик запроса на get
func Get(
	configuration *config.Configuration,
	ctx *web.Context,
	enc encoder.Encoder,
	mc *memCache.Cache,
	w http.ResponseWriter,
) (int, []byte) {
	// Подготовка структуры Request из GET параметров
	request := prepareRequest(configuration, ctx)
	if _, err := request.Validate(); err != nil {
		return resp.Fault(enc, err)
	}

	cacheId := request.Url  // Идентификатор кеша
	response := &Response{} // Ответ

	if data, found := mc.Get(cacheId); found {
		cache := data.(*Cache)    // Получаем структуру кеша
		response = cache.Response // Получаем структуру ответа
	} else {
		var err error

		// Формирует кеш ответа на запрос
		response, err = makeResponseCache(configuration, cacheId, request, mc)
		if err != nil {
			return resp.Fault(enc, err)
		}

		(&sync.Once{}).Do(onEvicted(configuration, mc))
	}

	// Пробрасываем список HTTP-заголовков в ответ
	for name, _ := range response.Response.Header {
		w.Header().Set(name, response.Response.Header.Get(name))
	}

	// Пробрасываем нативный ответ от запрашиваемого URL
	return response.Response.StatusCode, []byte(response.Body)
}

// Подготовка структуры Request из GET параметров
func prepareRequest(configuration *config.Configuration, ctx *web.Context) *Request {
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
func makeRequest(configuration *config.Configuration, request *Request) ([]byte, *http.Response, error) {
	/*
		// Получаем адрес прокси-сервера
		transport, err := getProxy(configuration, request)
		if err != nil {
			return nil, nil, err
		}
	*/

	// Структура HTTP-клиента
	client := http.Client{}

	/*
		// Если транспорт/прокси-сервер не определен, то это обычный запрос
		if transport != nil {
			// Используем прокси-сервер для запроса, если он указан
			// by https://stackoverflow.com/questions/14661511/setting-up-proxy-for-http-client
			//proxyUrl, err := url.Parse("http://163.172.215.220:80")
			//client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
			client.Transport = transport
		}
	*/

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
func makeResponseCache(configuration *config.Configuration, cacheId string, request *Request, mc *memCache.Cache) (*Response, error) {
	// Делает запрос по указанному URL
	body, resp, err := makeRequest(configuration, request)
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
	mc.Set(cacheId, cache, time.Duration(request.Inteval)*time.Second)

	fmt.Printf("makeResponseCache: %v\n", cacheId)

	return response, nil
}

// Возвращает адрес прокси-сервера для указанного запроса
// Список прокси-серверов ротируются по кольцу индивидуально для каждого запроса
func getProxy(configuration *config.Configuration, request *Request) (*http.Transport, error) {
	// Если в конфиге список серверов пуст, то на выход
	if len(configuration.Request.Proxy) == 0 {
		return nil, nil
	}

	// список-прокси серверов: http://spys.one/
	fmt.Printf("Proxy: %v\n", configuration.Request.Proxy)

	prxUrl := roundRobin(configuration.Request.Proxy, request.Proxy)

	proxyUrl, err := url.Parse(prxUrl)
	if err != nil {
		return nil, err
	}

	return &http.Transport{Proxy: http.ProxyURL(proxyUrl)}, nil
}

// Цикличная ротация строк из набора строк
func roundRobin(poll []string, current string) string {
	return "http://163.172.215.220:80"
}

func onEvicted(configuration *config.Configuration, mc *memCache.Cache) func() {
	return func() {
		// Обработчик callback на уборщик мусора из кеша (уборщик устаревшего кеша)
		mc.OnEvicted(func(key string, data interface{}) {
			fmt.Printf("OnEvicted: %v\n", key)

			cache := data.(*Cache)

			// Запихиваем старое значение в кеш, пока обновится собержимое кеша
			mc.Set(key, cache, time.Duration(cache.Request.Inteval)*time.Second)

			go func() {
				_, err := makeResponseCache(configuration, key, cache.Request, mc)
				if err != nil {
					panic(err.Error())
				}
			}()
		})
	}
}
