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
	Proxy   string        `valid:"-"`                       // Текущий прокси-сервер
}

// Конструктор структуры Request
func NewRequest(url string, inteval int, request *http.Request) *Request {
	return &Request{
		Url:     url,
		Inteval: inteval,
		Request: request,
		Proxy:   "",
	}
}

// Валидатор структуры Request
func (r *Request) Validate() (bool, error) {
	result, err := govalidator.ValidateStruct(r)

	if len(r.Proxy) != 0 && !govalidator.IsURL(r.Proxy) {
		return false, errors.New(fmt.Sprintf("Proxy to be set as proxy addres: %v", r.Proxy))
	}

	// Максимальный интервал опроса - 1 час
	if !govalidator.InRangeInt(r.Inteval, 1, (60 * 60)) {
		return false, errors.New(fmt.Sprintf("Range inteval to be min %v, max %v", 1, (60 * 60)))
	}

	return result, err
}
