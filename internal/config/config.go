package config

import (
	"flag"
	"os"

	"go.uber.org/zap"

	"github.com/dmad1989/urlcut/internal/logging"
)

const (
	defHost      = "localhost:8080"
	defShortHost = "http://localhost:8080"
	// defFileStorageDB = "/tmp/short-url-db.json"
	// defDBDSN=
)

var UserCtxKey = &ContextKey{"userId"}
var ErrorCtxKey = &ContextKey{"error"}

type ContextKey struct {
	name string
}

var conf = Config{
	url: defHost,
}

type Config struct {
	url           string
	shortAddress  string
	fileStoreName string
	dbConnName    string
}

func init() {
	flag.StringVar(&conf.url, "a", defHost, "server URL format host:port, :port")
	flag.StringVar(&conf.shortAddress, "b", defShortHost, "Address for short url")
	flag.StringVar(&conf.fileStoreName, "f", "", "file name for storage")
	flag.StringVar(&conf.dbConnName, "d", "", "database connection addres, format host=? port=? user=? password=? dbname=? sslmode=?")
}

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

func (c Config) GetURL() string {
	return c.url
}

func (c Config) GetShortAddress() string {
	return c.shortAddress
}

func (c Config) GetFileStoreName() string {
	return c.fileStoreName
}

func (c Config) GetDBConnName() string {
	return c.dbConnName
}
