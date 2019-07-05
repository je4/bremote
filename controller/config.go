package main

import (
	"github.com/BurntSushi/toml"
	"log"
)

type TemplatesInternalConfig struct {
	Name string
	File string
}

type TemplatesConfig struct {
	Folder     string
	DelimLeft  string
	DelimRight string
	Internal   []TemplatesInternalConfig
}

type Config struct {
	Logfile      string
	Loglevel     string
	InstanceName string
	Proxy        string
	CertPEM      string
	KeyPEM       string
	CaPEM        string
	HttpsCertPEM string
	HttpsKeyPEM  string
	HttpsAddr    string
	HttpStatic   string
	TLSAddr      string
	Templates    TemplatesConfig
}

func LoadConfig(filepath string) Config {
	var conf Config
	_, err := toml.DecodeFile(filepath, &conf)
	if err != nil {
		log.Fatalln("Error on loading config: ", err)
	}
	return conf
}
