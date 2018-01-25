package ping

import (
	"net/http"

	"github.com/martini-contrib/encoder"
)

// Обработчик запроса на ping
func Ping(enc encoder.Encoder) (int, []byte) {
	return http.StatusOK, encoder.Must(enc.Encode(NewResponse()))
}
