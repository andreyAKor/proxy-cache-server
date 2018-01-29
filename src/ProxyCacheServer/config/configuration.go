package config

type Configuration struct {
	App     AppConfiguration     // Информация о приложении
	Server  ServerConfiguration  // Данные для запуска сервера
	Request RequestConfiguration // Данные для запросов
}
