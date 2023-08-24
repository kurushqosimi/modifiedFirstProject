package main

import (
	"log"
	"main/internal/configs"
	"main/internal/handlers"
	"net/http"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}
func run() error {
	config, err := configs.InitConfigs()
	if err != nil {
		return err
	}
	address := config.Host + config.Port
	router := handlers.InitRoutes()
	srv := http.Server{
		Addr:    address,
		Handler: router,
	}
	err = srv.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}
