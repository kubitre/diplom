package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("everything okay! from slave node"))
	})
	if err := http.ListenAndServe(":9998", nil); err != nil {
		log.Panic(err)
	}
}
