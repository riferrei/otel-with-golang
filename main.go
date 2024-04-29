package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/hello", hello)
	http.ListenAndServe(":8888", router)
}

func hello(writer http.ResponseWriter, request *http.Request) {
	buildResponse(writer)
}

func buildResponse(writer http.ResponseWriter) {
	response := &struct {
		Message string `json:"Message"`
	}{
		Message: "Hello World!",
	}
	bytes, err := json.Marshal(response)
	if err != nil {
		writer.Write([]byte(err.Error()))
		writer.WriteHeader(http.StatusInternalServerError)
	}
	writer.Write(bytes)
	writer.Header().Add("Content-Type", "application/json")
}
