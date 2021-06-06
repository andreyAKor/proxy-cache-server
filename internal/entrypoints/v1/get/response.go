package get

import (
	"net/http"

	"github.com/asaskevich/govalidator"
)

// Структура ответа запроса get
type Response struct {
	Body     string         `valid:"required"` // Тело ответа
	Response *http.Response `valid:"-"`        // Содержимое HTTP-ответа
}

// Конструктор структуры Response
func NewResponse(body string, response *http.Response) *Response {
	return &Response{
		Body:     body,
		Response: response,
	}
}

// Валидатор структуры Response
func (r *Response) Validate() (bool, error) {
	return govalidator.ValidateStruct(r)
}
