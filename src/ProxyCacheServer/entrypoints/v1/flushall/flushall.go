package flushall

import (
	"net/http"

	"ProxyCacheServer/config"

	"github.com/codegangsta/martini-contrib/web"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/encoder"
	memCache "github.com/patrickmn/go-cache"
)

// Обработчик запроса на flushall
func Flushall(configuration *config.Configuration, ctx *web.Context, enc encoder.Encoder, mc *memCache.Cache, params martini.Params) (int, []byte) {

	// Delete all items from the cache.
	mc.Flush()

	// Структура ответа
	response := NewResponse("ok")
	if _, err := response.Validate(); err != nil {
		return SendError(enc, err)
	}

	return http.StatusOK, encoder.Must(enc.Encode(response))
}

// Формирует контент об ошибке
func SendError(enc encoder.Encoder, err error) (int, []byte) {
	return http.StatusBadRequest, encoder.Must(enc.Encode(map[string]string{
		"error": err.Error(),
	}))
}
