package main

import (
	"fmt"
	"log"
	"net/http"
)

const webPort = "80"

type Config struct {
}

func main() {

	//init Config
	app := Config{}
	log.Printf("strating server at port %s \n", webPort)
	//Create a http Server
	srv := http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}
	//Start the Http Server
	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}
