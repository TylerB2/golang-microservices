package main

import (
	"authentication-service/data"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"log"
	"net/http"
	"os"
	"time"
)

const webPort = "8081"

var counts int64

//Database Models Configuration
type Config struct {
	DB     *sql.DB
	Models data.Models
}

func main() {
	log.Println("Starting Authentication Service")
	//connect to DB
	conn := connectToDB()
	if conn == nil {
		log.Panic("Can't connect to postgress")
	}

	//Inittialize the config
	app := Config{
		DB:     conn,
		Models: data.New(conn),
	}

	//configure Server
	log.Printf("Starting server at Port %s", webPort)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	//Start the server
	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("Postgress not ready")
			counts++
		} else {
			log.Println("Connected to Postgress! ")
			return connection
		}

		if counts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("backing of for two seconds")
		time.Sleep(2 * time.Second)
		continue
	}
}
