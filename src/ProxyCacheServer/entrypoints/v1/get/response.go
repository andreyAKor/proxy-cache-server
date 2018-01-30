package get

import (
	"github.com/asaskevich/govalidator"
)

// Структура ответа запроса get
type Response struct {
	Body string `json:"body" valid:"required"` // Тело ответа
}

// Конструктор структуры Response
func NewResponse(body string) *Response {
	return &Response{
		Body: body,
	}
}

// Валидатор структуры Response
func (this *Response) Validate() (bool, error) {
	return govalidator.ValidateStruct(this)
}
