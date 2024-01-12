package config

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
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
