package main

import (
	"log"

	cfg "github.com/gasparian/go-api-template/internal/config"
	"github.com/gasparian/go-api-template/internal/server"
	"github.com/gasparian/go-api-template/internal/storage_driver"
)

var (
	configPath = cfg.GetAbsPath("./configs")
)

func main() {
	app := &server.App{}
	config, err := cfg.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}
	storage := storage_driver.NewKVDriver()
	app.Initialize(config.Application, config.CORS, storage)
	defer app.Logger.Sync() // Flushes buffer, if any
	app.Run(config.Server)
}
