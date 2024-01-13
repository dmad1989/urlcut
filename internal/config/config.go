package config

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/caarlos0/env/v6"
)

const (
	defHost = "localhost"
	defPort = 8080
)

var Conf = config{
	URL: netAddress{
		host: defHost,
		port: defPort},
	shortAddres: ""}

var confOs configOs

type configOs struct {
	Server_address string `env:"SERVER_ADDRESS"`
	ShortAddres    string `env:"BASE_URL"`
}

type config struct {
	URL         netAddress
	shortAddres string
}

type netAddress struct {
	host string
	port int
}

func init() {
	_ = flag.Value(&Conf.URL)
	flag.Var(&Conf.URL, "a", "server URL format host:port, :port")
	flag.StringVar(&Conf.shortAddres, "b", "", "Addres for short url")
}

func InitConfig() {
	flag.Parse()
	err := env.Parse(&confOs)
	if err != nil {
		log.Fatal(err)
	}
	if confOs.Server_address != "" {
		Conf.URL.Set(confOs.Server_address)
	}
	if confOs.ShortAddres != "" {
		fmt.Println("confOs.ShortAddres ", confOs.ShortAddres)
		Conf.SetShortAddress(confOs.ShortAddres)
	}
}

func (naddr netAddress) getHost() (res string) {
	res = naddr.host
	if res == "" {
		res = defHost
	}
	return
}

func (naddr *netAddress) String() string {
	return fmt.Sprintf("%s:%d", naddr.getHost(), naddr.port)
}

func (naddr *netAddress) Set(flagValue string) error {
	fmt.Println("flagValue ", flagValue)
	var sPort string
	var isFound bool
	var err error
	naddr.host, sPort, isFound = strings.Cut(flagValue, ":")
	if !isFound {
		panic("On declared -a not url param")
	}
	naddr.port, err = strconv.Atoi(sPort)
	if err != nil {
		return err
	}
	return err
}

func (c config) GetShortAddress() (res string) {
	res = c.shortAddres
	if res == "" {
		res = c.URL.String()
	}
	return
}

func (c *config) SetShortAddress(newVal string) {
	if newVal != "" {
		c.shortAddres = newVal
	}
}
