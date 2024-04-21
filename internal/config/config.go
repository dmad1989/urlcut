// Модуль config инициализирует конфигурацию при запуске сервера.
// Данные инициализируются согласно следующего приоритета:
// из переменной окружения, если ее нет - из флага, указанного при запуске, если его нет - значение по умолчанию.
package config

import (
	"flag"
	"os"

	"go.uber.org/zap"

	"github.com/dmad1989/urlcut/internal/logging"
)

// Значения по умолчанию.
const (
	defHost      = "localhost:8080"
	defShortHost = "http://localhost:8080"
)

// Ключи для данных передающихся в контексте.
var (
	UserCtxKey  = &ContextKey{"userId"} // ID пользователя
	ErrorCtxKey = &ContextKey{"error"}  // ошибка
)

type ContextKey struct {
	name string
}

var conf = Config{
	url: defHost,
}

// Config хранит параметры для запуска сервера.
type Config struct {
	// url - адрес по которому будет доступен сервер.
	url string
	// shortAddress - адрес, который будет в сокращении.
	shortAddress string
	// fileStoreName - имя json - файла, для хранения сокращенных URL.
	fileStoreName string
	//dbConnName - DSN к PostgresSQL
	dbConnName string
}

// Инициализация конфигурации значениями флага или по умолчанию.
func init() {
	flag.StringVar(&conf.url, "a", defHost, "server URL format host:port, :port")
	flag.StringVar(&conf.shortAddress, "b", defShortHost, "Address for short url")
	flag.StringVar(&conf.fileStoreName, "f", "", "file name for storage")
	flag.StringVar(&conf.dbConnName, "d", "", "database connection addres, format host=? port=? user=? password=? dbname=? sslmode=?")
}

// ParseConfig - запускает парсинг флагов и анализирует переменные окружения.
func ParseConfig() Config {
	flag.Parse()
	if os.Getenv("SERVER_ADDRESS") != "" {
		conf.url = os.Getenv("SERVER_ADDRESS")
	}

	if os.Getenv("BASE_URL") != "" {
		conf.shortAddress = os.Getenv("BASE_URL")
	}

	if os.Getenv("FILE_STORAGE_PATH") != "" {
		conf.fileStoreName = os.Getenv("FILE_STORAGE_PATH")
	}

	if os.Getenv("DATABASE_DSN") != "" {
		conf.dbConnName = os.Getenv("DATABASE_DSN")
	}

	logging.Log.Debugw("starting config ",
		zap.String("URL", conf.url),
		zap.String("shortAddress", conf.shortAddress),
		zap.String("fileStoreName", conf.fileStoreName),
		zap.String("dbConnName", conf.dbConnName))
	return conf
}

// GetURL - получить адрес по которому будет запущен сервер.
func (c Config) GetURL() string {
	return c.url
}

// GetShortAddress - получить адрес, который будет в ответе с сокращением.
func (c Config) GetShortAddress() string {
	return c.shortAddress
}

// GetFileStoreName - получить путь к файлу с сокращениями
func (c Config) GetFileStoreName() string {
	return c.fileStoreName
}

// GetDBConnName - получить DSN к DB
func (c Config) GetDBConnName() string {
	return c.dbConnName
}
