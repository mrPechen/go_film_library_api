package main

import (
	"fmt"
	"log"
	"net/http"
	"src/src/database"
	"src/src/routers"
)

func main() {

	pool, PoolErr := database.DBConnect()
	if PoolErr != nil {
		log.Fatalf("Unable to connect to database: %v", PoolErr)
	}
	defer pool.Close()
	if err := database.CreateTables(pool); err != nil {
		log.Fatalf("Unable to create tables: %v", err)
	}

	router := http.NewServeMux()

	routers.Handlers(pool, router)

	fmt.Println("Sever started")
	log.Fatal(http.ListenAndServe(":8080", router))
}
