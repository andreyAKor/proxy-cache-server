package flushall

import (
	"net/http"

	"github.com/andreyAKor/proxy-cache-server/internal/config"
	resp "github.com/andreyAKor/proxy-cache-server/internal/entrypoints/v1/response"

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
		return resp.Fault(enc, err)
	}

	return http.StatusOK, encoder.Must(enc.Encode(response))
}
