package ping

// Структура ответа запроса ping
type Response struct {
	Pong bool `json:"pong" valid:"-"`
}

// Конструктор структуры Response
func NewResponse() *Response {
	return &Response{
		Pong: true,
	}
}
