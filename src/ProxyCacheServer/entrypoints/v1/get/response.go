package get

import (
	"github.com/asaskevich/govalidator"
)

// Структура ответа запроса get
type Response struct {
	Value string `json:"value" valid:"required"` // Значение
}

// Конструктор структуры Response
func NewResponse(value string) *Response {
	return &Response{
		Value: value,
	}
}

// Валидатор структуры Response
func (this *Response) Validate() (bool, error) {
	result, err := govalidator.ValidateStruct(this)

	return result, err
}
