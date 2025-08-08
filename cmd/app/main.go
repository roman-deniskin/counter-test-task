package main

import (
	"log"
	"net/http"

	"counter-test-task/internal/db"
	"counter-test-task/internal/handler"
	"counter-test-task/internal/router"
	"counter-test-task/internal/service"
)

func main() {
	dbConn := db.MustConnect()
	srv := service.New(dbConn)
	h := handler.New(srv)

	r := router.New(h)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
