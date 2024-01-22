package config

import (
	"flag"
	"os"
)

const (
	defHost      = "localhost:8080"
	defShortHost = "http://localhost:8080"
)

var conf = Config{
	url:          defHost,
	shortAddress: ""}

type Config struct {
	url          string
	shortAddress string
}

func init() {
	flag.StringVar(&conf.url, "a", defHost, "server URL format host:port, :port")
	flag.StringVar(&conf.shortAddress, "b", defShortHost, "Address for short url")
}

func ParseConfig() Config {
	flag.Parse()

	if os.Getenv("SERVER_ADDRESS") != "" {
		conf.url = os.Getenv("SERVER_ADDRESS")
	}

	if os.Getenv("BASE_URL") != "" {
		conf.shortAddress = os.Getenv("BASE_URL")
	}

	return conf
}

func (c Config) GetURL() string {
	return c.url
}

func (c Config) GetShortAddress() string {
	return c.shortAddress
}
