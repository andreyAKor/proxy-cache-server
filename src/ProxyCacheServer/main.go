/*
http://localhost:8383/v1/ping/
http://localhost:8383/v1/get/
http://localhost:8383/v1/get/?url=https://www.google.ru/
http://localhost:8383/v1/get/?url=https://api.bitfinex.com/v1/pubticker/sntusd
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"ProxyCacheServer/config"
	"ProxyCacheServer/entrypoints/v1/get"
	"ProxyCacheServer/entrypoints/v1/ping"

	"github.com/codegangsta/martini-contrib/web"
	"github.com/go-martini/martini"
	"github.com/kardianos/service"
	"github.com/martini-contrib/encoder"
	memCache "github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
)

// Структура приложения
type App struct {
	configuration *config.Configuration
	webServer     *martini.ClassicMartini
}

// Обработчик приложения
func (app *App) Run() {
	// Указываем свой хост и порт и слушаем его
	app.webServer.RunOnAddr(app.configuration.Server.Host + ":" + strconv.Itoa(app.configuration.Server.Port))
}

// Инициализация системного лога
var logger service.Logger

// Структура программы
type program struct {
	exit    chan struct{}
	service service.Service
	app     *App
}

// Обработчик старта сервиса
func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	logger.Info("Start")

	go p.run()
	return nil
}

// Обработчик программы сервиса
func (p *program) run() {
	logger.Info("Runnig ", p.app.configuration.App.DisplayName)

	defer func() {
		// Смотрим наличие менеджеров сервисов в ОС
		if service.ChosenSystem() != nil {
			if service.Interactive() {
				p.Stop(p.service)
			} else {
				p.service.Stop()
			}
		}
	}()

	// Запускаем обработчик приложения
	p.app.Run()
}

// Обработчик остановки сервиса
func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	logger.Info("Stop")

	return nil
}

// Инициализация кешера
func InitCache() (*memCache.Cache, error) {
	// Create a cache with a default expiration time of 1 hours, and which purges expired items every 10 minutes
	cache := memCache.New(time.Second, time.Second)

	return cache, nil
}

// Инициализация веб-сервера
func InitWebServer(configuration *config.Configuration) (*martini.ClassicMartini, error) {
	// Инициализация сервера Мартини
	m := martini.Classic()

	// Настройка "middleware"
	// Прикручиваем контекст по работе c "сырым" веб-запросами
	// by https://github.com/martini-contrib/web, https://github.com/hoisie/web
	m.Use(web.ContextWithCookieSecret(""))

	// Сервис для представления данных в нескольких форматах и взаимодействия с контентом
	m.Use(func(c martini.Context, w http.ResponseWriter, r *http.Request) {
		c.MapTo(encoder.JsonEncoder{}, (*encoder.Encoder)(nil))
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	})

	// Роутинг запросов
	m.Get("/v1/ping/", ping.Ping)
	m.Get("/v1/get/(.*)", get.Get)

	return m, nil
}

// Инициализация конфига приложения
func InitConfiguration() (*config.Configuration, error) {
	// Имя файла yml-конфига
	viper.SetConfigName("ProxyCacheServer")

	// По умолчанию конфиг лежит тамже, где и приложение
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/ProxyCacheServer")

	// Структура конфига
	var configuration config.Configuration

	// Читаем конфиг-файл
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("Error reading config file, %s", err)
	}

	if err := viper.Unmarshal(&configuration); err != nil {
		return nil, fmt.Errorf("Unable to decode into struct, %v", err)
	}

	return &configuration, nil
}

// Инициализация инстанса сервиса
func InitService(app *App, svcFlag *string) error {
	// Структура программы
	prg := &program{
		exit: make(chan struct{}),
		app:  app,
	}

	// Конфиг сервиса
	svcConfig := &service.Config{
		Name:        app.configuration.App.Name,
		DisplayName: app.configuration.App.DisplayName,
		Description: app.configuration.App.Description,
	}

	// Создание экземпляра сервиса
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	prg.service = s

	// Инициализация системного логгера
	errs := make(chan error, 5)
	logger, err = s.Logger(errs)
	if err != nil {
		log.Fatal(err)
	}

	// Вывод лога ошибок в консоль терминала
	go func() {
		for {
			err := <-errs
			if err != nil {
				log.Print(err)
			}
		}
	}()

	// Управление сервисом
	if len(*svcFlag) != 0 {
		err := service.Control(s, *svcFlag)
		if err != nil {
			log.Printf("Valid actions: %q\n", service.ControlAction)
			log.Fatal(err)
		}

		return nil
	}

	// Запуск обработчика сервиса
	err = s.Run()
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

// Точка входа
func main() {
	// Инициализация конфига приложения
	configuration, err := InitConfiguration()
	if err != nil {
		log.Fatal(err)
	}

	// Инициализация веб-сервера
	webServer, err := InitWebServer(configuration)
	if err != nil {
		log.Fatal(err)
	}

	// Инициализация кешера
	cache, err := InitCache()
	if err != nil {
		log.Fatal(err)
	}

	// Сервис для кешера
	webServer.Map(cache)

	// Маппинг конфига
	webServer.Map(configuration)

	// Структура приложения
	app := &App{
		configuration: configuration,
		webServer:     webServer,
	}

	// Смотрим наличие менеджеров сервисов в ОС
	// Если хоть что-то есть, то запускаем приложение через системный менеджер сервисов
	// иначе приложение будет работать как обычная программа
	if service.ChosenSystem() != nil {
		fmt.Printf("Service system is available: %v\n", service.AvailableSystems())

		svcFlag := flag.String("service", "", "Control the system service.")
		flag.Parse()

		// Инициализация инстанса сервиса
		if err := InitService(app, svcFlag); err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Printf("Service system is not found\n")

		// Запускаем обработчик приложения
		app.Run()
	}
}
