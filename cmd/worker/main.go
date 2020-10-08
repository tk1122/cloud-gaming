package main

import (
	"github.com/tk1122/cloud-gaming/pkg/worker"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	f, err := os.OpenFile("all.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		_ = f.Close()
	}()

	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	log.Println("http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", worker.Router))
	//fmt.Println("https://localhost:8000")
	//log.Fatal(http.ListenAndServeTLS(":8000", "cert.pem", "key.pem", worker.Router))
}
