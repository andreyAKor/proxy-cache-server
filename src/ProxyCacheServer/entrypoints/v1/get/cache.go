package get

import (
	"github.com/asaskevich/govalidator"
)

// Структура для кеширования данных запроса и ответа
type Cache struct {
	Request  *Request  `valid:"-"` // Данные запроса
	Response *Response `valid:"-"` // Данные ответа
}

// Конструктор структуры Cache
func NewCache(request *Request, response *Response) *Cache {
	return &Cache{
		Request:  request,
		Response: response,
	}
}

// Валидатор структуры Cache
func (this *Cache) Validate() (bool, error) {
	return govalidator.ValidateStruct(this)
}
