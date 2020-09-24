package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/tk1122/cloud-gaming/pkg/worker/emulator"
	"github.com/tk1122/cloud-gaming/pkg/worker/webrtc"
	"image"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var webRTC *webrtc.WebRTC
var width = 256
var height = 240

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	genPem()
}

func startGame(path string, imageChannel chan *image.RGBA, inputChannel chan string) {
	director := emulator.NewDirector(imageChannel, inputChannel)
	director.Start([]string{path})
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", getWeb).Methods("GET")
	router.HandleFunc("/ws", getWs).Methods("GET")
	//webRTC = webrtc.NewWebRTC()

	fmt.Println("http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", router))
	//log.Fatal(http.ListenAndServeTLS(":8000", "cert.pem", "key.pem", router))
}

//func postSession(w http.ResponseWriter, r *http.Request) {
//	bs, err := ioutil.ReadAll(r.Body)
//	if err != nil {
//		log.Fatal(err)
//	}
//	r.Body.Close()
//
//	localSession, err := webRTC.StartClient(string(bs), width, height)
//	if err != nil {
//		log.Fatalln(err)
//	}
//
//	imageChannel := make(chan *image.RGBA, 2)
//	go screenshotLoop(imageChannel)
//	go startGame("games/supermariobros.rom", imageChannel, webRTC.InputChannel)
//
//	w.Write([]byte(localSession))
//
//}
//
//func randomImage(width, height int) *image.RGBA {
//	img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{width, height}})
//	for x := 0; x < width; x++ {
//		for y := 0; y < height; y++ {
//			img.Set(x, y, color.RGBA{uint8(rand.Int31n(0xff)), uint8(rand.Int31n(0xff)), uint8(rand.Int31n(0xff)), 0xff - 1})
//		}
//	}
//
//	return img
//}
//
//func screenshotLoop(imageChannel chan *image.RGBA) {
//	for image := range imageChannel {
//		if webRTC.IsConnected() {
//			//rgbaImg := randomImage(width, height)
//			yuv := screenshot.RgbaToYuv(image)
//			webRTC.ImageChannel <- yuv
//		}
//	}
//}
