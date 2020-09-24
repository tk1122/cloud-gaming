package main

import (
	"fmt"
	"github.com/tk1122/cloud-gaming/pkg/worker"
	"log"
	"net/http"
)

func main() {
	fmt.Println("http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", worker.Router))
	//log.Fatal(http.ListenAndServeTLS(":8000", "cert.pem", "key.pem", router))
}
