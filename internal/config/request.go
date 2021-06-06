package config

// Данные для запросов
type RequestConfiguration struct {
	Interval int      // Интервал (периодичность) опроса URL-адреса в секундах
	Proxy    []string // Список прокси-серверов
}
