package config

import (
	"flag"
	"os"
)

const (
	defHost      = "localhost:8080"
	defShortHost = "http://localhost:8080"
)

var Conf = config{
	URL:          defHost,
	ShortAddress: ""}

type config struct {
	URL          string
	ShortAddress string
}

func init() {
	flag.StringVar(&Conf.URL, "a", defHost, "server URL format host:port, :port")
	flag.StringVar(&Conf.ShortAddress, "b", defShortHost, "Address for short url")
}

func InitConfig() {
	flag.Parse()

	if os.Getenv("SERVER_ADDRESS") != "" {
		Conf.URL = os.Getenv("SERVER_ADDRESS")
	}

	if os.Getenv("BASE_URL") != "" {
		Conf.ShortAddress = os.Getenv("BASE_URL")
	}

	// if Conf.ShortAddress == "" {
	// 	Conf.ShortAddress = Conf.URL
	// }
}
