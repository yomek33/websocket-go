package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const port = 8080

func main() {
	route := mux.NewRouter()
	Routes(route)

	http.Handle("/", route)

	log.Fatal(http.ListenAndServe(":8080", nil))
	log.Println("starting server on port", port)

}
