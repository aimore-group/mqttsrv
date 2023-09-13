// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 mochi-mqtt, mochi-co
// SPDX-FileContributor: mochi-co

package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aimore-group/mqttsrv"
	"github.com/aimore-group/mqttsrv/config"
	"github.com/aimore-group/mqttsrv/hooks/auth"
	"github.com/aimore-group/mqttsrv/listeners"
)

func newTlsConfig(pemFile string, keyFile string, caFile string) *tls.Config {
	pem, err := tls.LoadX509KeyPair(pemFile, keyFile)
	if err != nil {
		log.Fatalf("[ERROR] tls serve read cert file:%s err:%s\n", pemFile, err.Error())
	}
	ca, err := os.ReadFile(caFile)
	if err != nil {
		log.Fatalf("[ERROR] tls serve read ca file failed. err:%s\n", err.Error())
	}

	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(ca)
	if !caPool.AppendCertsFromPEM(ca) {
		log.Fatalf("[ERROR] tls serve append cert pool failed.\n")
	}
	return &tls.Config{
		Certificates:       []tls.Certificate{pem},
		ClientCAs:          caPool,
		InsecureSkipVerify: false,
		ClientAuth:         tls.RequireAndVerifyClientCert,
	}
}

func main() {
	configFile := flag.String("c", "config.toml", "The config file path for mqtt server.")
	flag.Parse()

	conf := config.FromFile(*configFile)
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()

	server := mqttsrv.New(&conf.Options)
	_ = server.AddHook(new(auth.AllowHook), nil)

	if len(conf.TcpAddr) > 0 {
		tcp := listeners.NewTCP("t1", conf.TcpAddr, nil)
		err := server.AddListener(tcp)
		if err != nil {
			log.Fatal(err)
		}
	}

	if len(conf.TlsAddr) > 0 {
		tlsConfig := newTlsConfig(conf.TlsCert, conf.TlsKey, conf.TlsCa)
		tlsTcp := listeners.NewTCP("tls1", conf.TlsAddr, &listeners.Config{
			TLSConfig: tlsConfig,
		})
		err := server.AddListener(tlsTcp)
		if err != nil {
			log.Fatal(err)
		}
	}

	if len(conf.WsAddr) > 0 {
		ws := listeners.NewWebsocket("ws1", conf.WsAddr, nil)
		err := server.AddListener(ws)
		if err != nil {
			log.Fatal(err)
		}
	}

	if len(conf.StatsAddr) > 0 {
		stats := listeners.NewHTTPStats("stats", conf.StatsAddr, nil, server.Info)
		err := server.AddListener(stats)
		if err != nil {
			log.Fatal(err)
		}
	}

	go func() {
		err := server.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	<-done
	server.Log.Warn("caught signal, stopping...")
	_ = server.Close()
	server.Log.Info("main.go finished")
}
