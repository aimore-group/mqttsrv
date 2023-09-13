package config

import (
	"log"
	"os"

	"github.com/aimore-group/mqttsrv"
	"github.com/naoina/toml"
)

type Config struct {
	TlsAddr   string
	TcpAddr   string
	WsAddr    string
	StatsAddr string

	TlsCa   string
	TlsCert string
	TlsKey  string
	Options mqttsrv.Options
}

func FromFile(file string) *Config {
	var conf Config
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	if err := toml.NewDecoder(f).Decode(&conf); err != nil {
		log.Fatal(err)
	}
	return &conf
}
