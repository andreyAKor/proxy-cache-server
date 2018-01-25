package get

import (
	"errors"
	"fmt"

	"github.com/asaskevich/govalidator"
)

// Структура запроса для запроса get
type Request struct {
	Url   string `valid:"required,length(1|255)"` // Опрашиваемый URL-адрес
	Limit int    `valid:"required"`               // -
}

// Конструктор структуры Request
func NewRequest(url string) *Request {
	return &Request{
		Url:   url,
		Limit: 10,
	}
}

// Валидатор структуры Request
func (this *Request) Validate() (bool, error) {
	result, err := govalidator.ValidateStruct(this)

	if !govalidator.InRangeInt(this.Limit, 1, (60 * 60)) {
		return false, errors.New(fmt.Sprintf("Range limit to be min %v, max %v", 1, (60 * 60)))
	}

	return result, err
}
