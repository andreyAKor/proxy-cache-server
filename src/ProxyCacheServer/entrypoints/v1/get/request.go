package get

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/asaskevich/govalidator"
)

// Структура запроса для запроса get
type Request struct {
	Url     string `valid:"required,length(1|2048)"` // Опрашиваемый URL-адрес
	Inteval int    `valid:"required"`                // Интервал (периодичность) опроса URL-адреса в секундах

	// HTTP содержимое клиента
	UserAgent         string         `valid:"required"` // UserAgent браузера
	Method            string         `valid:"required"` // Метод HTTP-запроса
	Proto             string         `valid:"required"` // Версия HTTP-протокола
	Referer           string         `valid:"-"`        // Откуда пришёл браузер
	BasicAuthUsername string         `valid:"-"`        // Имя пользователя для basic-авторизации
	BasicAuthPassword string         `valid:"-"`        // Пароль для basic-авторизации
	Cookies           []*http.Cookie `valid:"-"`        // Список cookie-данных
	Header            http.Header    `valid:"-"`        // Список HTTP-заголовков
}

// Конструктор структуры Request
func NewRequest(url string, inteval int, userAgent, method, proto string) *Request {
	return &Request{
		Url:     url,
		Inteval: inteval,

		// HTTP содержимое клиента
		UserAgent: userAgent,
		Method:    method,
		Proto:     proto,
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
