package main

import (
	"github.com/BurntSushi/toml"
	"log"
)

type Config struct {
	Logfile      string
	Loglevel     string
	InstanceName string
	CertPEM      string
	KeyPEM       string
	CaPEM        string
	TLSAddr      string
	WebRoot      string
	KVDBFile     string
	NTPHost      string
}

func LoadConfig(filepath string) Config {
	var conf Config
	_, err := toml.DecodeFile(filepath, &conf)
	if err != nil {
		log.Fatalln("Error on loading config: ", err)
	}
	return conf
}
