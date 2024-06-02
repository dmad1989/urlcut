// Package config инициализирует конфигурацию при запуске сервера.
// Данные инициализируются согласно следующего приоритета:
// из переменной окружения, если ее нет - из флага, указанного при запуске, если его нет - значение по умолчанию.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

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
	TokenCtxKey = &ContextKey{"token"}  // токен
)

// ContextKey реализует ключ для значения в контексте.
type ContextKey struct {
	name string
}

// Config хранит параметры для запуска сервера.
type Config struct {
	URL           string `json:"server_address"`
	ShortAddress  string `json:"base_url"`
	FileStoreName string `json:"file_storage_path"`
	DBConnName    string `json:"database_dsn"`
	EnableHTTPS   bool   `json:"enable_https"`
	TrustedSubnet string `json:"trusted_subnet"`
	filePath      string
}

// ParseConfig - запускает парсинг флагов и анализирует переменные окружения.
func ParseConfig() (conf Config, err error) {
	conf = Config{
		URL: defHost,
	}
	conf.initFlags()

	if os.Getenv("SERVER_ADDRESS") != "" {
		conf.URL = os.Getenv("SERVER_ADDRESS")
	}

	if os.Getenv("BASE_URL") != "" {
		conf.ShortAddress = os.Getenv("BASE_URL")
	}

	if os.Getenv("FILE_STORAGE_PATH") != "" {
		conf.FileStoreName = os.Getenv("FILE_STORAGE_PATH")
	}

	if os.Getenv("DATABASE_DSN") != "" {
		conf.DBConnName = os.Getenv("DATABASE_DSN")
	}

	if os.Getenv("TRUSTED_SUBNET") != "" {
		conf.DBConnName = os.Getenv("TRUSTED_SUBNET")
	}

	if os.Getenv("ENABLE_HTTPS") != "" {
		b, err := strconv.ParseBool(os.Getenv("ENABLE_HTTPS"))
		if err != nil {
			logging.Log.Errorw("fails to read ENABLE_HTTPS", zap.Error(err))
		}
		conf.EnableHTTPS = b
	}

	if p, b := os.LookupEnv("CONFIG"); b {
		conf.filePath = p
	}

	if err = conf.loadFromFile(); err != nil {
		err = fmt.Errorf("config: loadfromFile: %w", err)
	}

	logging.Log.Infow("starting config ",
		zap.String("URL", conf.URL),
		zap.String("shortAddress", conf.ShortAddress),
		zap.String("fileStoreName", conf.FileStoreName),
		zap.String("dbConnName", conf.DBConnName),
		zap.Bool("ENABLE_HTTPS", conf.EnableHTTPS),
		zap.String("CONFIG", conf.filePath),
		zap.String("trusted IPs", conf.TrustedSubnet),
		zap.Error(err),
	)
	return conf, err
}

// GetURL - получить адрес по которому будет запущен сервер.
func (c Config) GetURL() string {
	return c.URL
}

// GetShortAddress - получить адрес, который будет в ответе с сокращением.
func (c Config) GetShortAddress() string {
	return c.ShortAddress
}

// GetFileStoreName - получить путь к файлу с сокращениями
func (c Config) GetFileStoreName() string {
	return c.FileStoreName
}

// GetDBConnName - получить DSN к DB
func (c Config) GetDBConnName() string {
	return c.DBConnName
}

// GetEnableHTTPS - запустить https-сервер
func (c Config) GetEnableHTTPS() bool {
	return c.EnableHTTPS
}

// GetTrustedSubnet - строковое представление бесклассовой адресации (CIDR)
func (c Config) GetTrustedSubnet() string {
	return c.TrustedSubnet
}

func (c *Config) initFlags() {
	flag.StringVar(&c.URL, "a", defHost, "server URL format host:port, :port")
	flag.StringVar(&c.ShortAddress, "b", defShortHost, "Address for short url")
	flag.StringVar(&c.FileStoreName, "f", "", "file name for storage")
	flag.StringVar(&c.DBConnName, "d", "", "database connection addres, format host=? port=? user=? password=? dbname=? sslmode=?")
	flag.BoolVar(&c.EnableHTTPS, "s", false, "true for htts server start")
	flag.StringVar(&c.filePath, "c", "", "path to config json file")
	flag.StringVar(&c.TrustedSubnet, "t", "", "trusted CIDR")
	flag.Parse()
}

func (c *Config) loadFromFile() error {
	if c.filePath == "" {
		return nil
	}

	b, err := os.ReadFile(c.filePath)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	var jConf Config
	if err = json.Unmarshal(b, &jConf); err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	c.URL = notEmptyVal(c.URL, jConf.URL)
	c.ShortAddress = notEmptyVal(c.ShortAddress, jConf.ShortAddress)
	c.FileStoreName = notEmptyVal(c.FileStoreName, jConf.FileStoreName)
	c.DBConnName = notEmptyVal(c.DBConnName, jConf.DBConnName)
	c.EnableHTTPS = notEmptyVal(c.EnableHTTPS, jConf.EnableHTTPS)
	c.TrustedSubnet = notEmptyVal(c.TrustedSubnet, jConf.TrustedSubnet)
	return nil
}
func notEmptyVal[T comparable](c T, j T) T {
	var zero T
	if c == zero {
		return j
	}
	return c
}
