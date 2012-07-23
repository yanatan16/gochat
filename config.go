package gochat

// Configuration parameters for gochat
// Set these with SetConfig before instantiating anything
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

var Cfg *Config

func TheConfig() *Config {
	return Cfg
}

func init() {
	fillWithDefaults()
}

func fillWithDefaults() {
	Cfg = &Config{
		WsAddr:      "127.0.0.1",
		WsPort:      8001,
		DbAddr:      "tcp:127.0.0.1:6379",
		DbDb:        0,
		DbPassword:  "",
		SubAddr:     "tcp:127.0.0.1:6379",
		SubDb:       0,
		SubPassword: "",
	}
}
