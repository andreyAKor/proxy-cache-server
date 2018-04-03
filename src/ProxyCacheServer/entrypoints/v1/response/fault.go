package response

import (
	"net/http"

	"github.com/martini-contrib/encoder"
)

// Формирует контент об ошибке
func Fault(enc encoder.Encoder, err error) (int, []byte) {
	return http.StatusBadRequest, encoder.Must(enc.Encode(map[string]string{
		"error": err.Error(),
	}))
}
