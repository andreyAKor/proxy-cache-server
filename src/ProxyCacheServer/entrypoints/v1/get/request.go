package get

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/asaskevich/govalidator"
)

// Структура запроса для запроса get
type Request struct {
	Url     string        `valid:"required,length(1|2048)"` // Опрашиваемый URL-адрес
	Inteval int           `valid:"required"`                // Интервал (периодичность) опроса URL-адреса в секундах
	Request *http.Request `valid:"-"`                       // Содержимое HTTP-запроса
}

// Конструктор структуры Request
func NewRequest(url string, inteval int, request *http.Request) *Request {
	return &Request{
		Url:     url,
		Inteval: inteval,
		Request: request,
	}
}

// Валидатор структуры Request
func (this *Request) Validate() (bool, error) {
	result, err := govalidator.ValidateStruct(this)

	// Максимальный интервал опроса - 1 час
	if !govalidator.InRangeInt(this.Inteval, 1, (60 * 60)) {
		return false, errors.New(fmt.Sprintf("Range inteval to be min %v, max %v", 1, (60 * 60)))
	}

	return result, err
}
