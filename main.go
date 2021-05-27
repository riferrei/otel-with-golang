package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {

	// Register the API handler and starts the app
	router := mux.NewRouter()
	router.HandleFunc("/hello", hello)
	http.ListenAndServe(":8888", router)

}

func hello(writer http.ResponseWriter, request *http.Request) {

	response := buildResponse(writer)

	if response.isValid() {
		log.Print("The response is valid")
	}

}

func buildResponse(writer http.ResponseWriter) Response {

	writer.WriteHeader(http.StatusOK)
	writer.Header().Add("Content-Type",
		"application/json")

	bytes, _ := json.Marshal("Hello World")
	writer.Write(bytes)
	return Response{}

}

// Response struct
type Response struct {
}

func (r Response) isValid() bool {
	return true
}
