package main

import (
	"counter-test-task/internal/db"
	"counter-test-task/internal/handler"
	"log"
	"net/http"
)

func main() {
	dbConn := db.MustConnect()
	h := handler.New(dbConn)
	http.HandleFunc("/", h.Hello)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
