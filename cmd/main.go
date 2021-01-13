package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	router.Handle("/", http.HandlerFunc(handle))
	log.Fatal(http.ListenAndServe(":8000", router))
}

func handle(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hello, World!")
}
