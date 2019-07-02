package main

import (
	"github.com/BurntSushi/toml"
	"log"
)

type Config struct {
	Logfile       string
	Loglevel      string
	InstanceName  string
	Proxy         string
	CertPEM       string
	KeyPEM        string
	CaPEM         string
	HttpsCertPEM  string
	HttpsKeyPEM   string
	HttpsAddr     string
	HttpStatic    string
	HttpTemplates string
}

func LoadConfig(filepath string) Config {
	var conf Config
	_, err := toml.DecodeFile(filepath, &conf)
	if err != nil {
		log.Fatalln("Error on loading config: ", err)
	}
	return conf
}
