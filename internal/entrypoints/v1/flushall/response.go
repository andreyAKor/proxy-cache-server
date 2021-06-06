package flushall

import (
	"github.com/asaskevich/govalidator"
)

// Структура ответа запроса flushall
type Response struct {
	Status string `json:"status" valid:"required"` // Статус ответа
}

// Конструктор структуры Response
func NewResponse(status string) *Response {
	return &Response{
		Status: status,
	}
}

// Валидатор структуры Response
func (this *Response) Validate() (bool, error) {
	return govalidator.ValidateStruct(this)
}
