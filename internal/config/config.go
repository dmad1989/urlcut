package config

import (
	"flag"
	"os"

	"github.com/dmad1989/urlcut/internal/logging"
	"go.uber.org/zap"
)

const (
	defHost          = "localhost:8080"
	defShortHost     = "http://localhost:8080"
	defFileStorageDB = "/tmp/short-url-db.json"
)

var conf = Config{
	url:          defHost,
	shortAddress: ""}

type Config struct {
	url           string
	shortAddress  string
	fileStoreName string
}

func init() {
	flag.StringVar(&conf.url, "a", defHost, "server URL format host:port, :port")
	flag.StringVar(&conf.shortAddress, "b", defShortHost, "Address for short url")
	flag.StringVar(&conf.fileStoreName, "f", defFileStorageDB, "file name for storage")
}

func ParseConfig() Config {
	flag.Parse()
	defer logging.Log.Sync()
	if os.Getenv("SERVER_ADDRESS") != "" {
		conf.url = os.Getenv("SERVER_ADDRESS")
	}

	if os.Getenv("BASE_URL") != "" {
		conf.shortAddress = os.Getenv("BASE_URL")
	}

	if os.Getenv("FILE_STORAGE_PATH") != "" {
		conf.fileStoreName = os.Getenv("FILE_STORAGE_PATH")
	}
	logging.Log.Debugw("starting config ",
		zap.String("URL", conf.url),
		zap.String("shortAddress", conf.shortAddress),
		zap.String("fileStoreName", conf.fileStoreName))
	return conf
}

func (c Config) GetURL() string {
	return c.url
}

func (c Config) GetShortAddress() string {
	return c.shortAddress
}

func (c Config) GetFileStoreName() string {
	return c.fileStoreName
}
