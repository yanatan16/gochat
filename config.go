package gochat

import (
	"encoding/json"

	"io/ioutil"
	"log"
)

type Config struct {
	// Websocket Configuration
	WsAddr string
	WsPort int

	// Redis Database
	DbAddr     string
	DbDb       int
	DbPassword string

	// Redis subscription database
	SubAddr     string
	SubDb       int
	SubPassword string
}

var Cfg Config

func init() {
	ReadConfig()
}

func ReadConfig() {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatalln("Could not open configuration file!", err)
	}
	err = json.Unmarshal(data, &Cfg)
	if err != nil {
		log.Fatalln("Could not unmarshal json configuration!", err)
	}

	log.Println("Read Config:", Cfg)
}
