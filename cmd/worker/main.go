package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/poi5305/go-yuv2webRTC/screenshot"
	"github.com/tk1122/cloud-gaming/pkg/worker/emulator"
	"github.com/tk1122/cloud-gaming/pkg/worker/webrtc"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
)

var webRTC *webrtc.WebRTC
var width = 256
var height = 240

func init() {
}

func startGame(path string, imageChannel chan *image.RGBA, inputChannel chan string) {
	emulator.Run([]string{path}, imageChannel, inputChannel)
}

func main() {
	fmt.Println("http://localhost:8000")
	webRTC = webrtc.NewWebRTC()

	router := mux.NewRouter()
	router.HandleFunc("/", getWeb).Methods("GET")
	router.HandleFunc("/session", postSession).Methods("POST")

	http.ListenAndServe(":8000", router)
}

func getWeb(w http.ResponseWriter, r *http.Request) {
	bs, err := ioutil.ReadFile("./web/index.html")
	if err != nil {
		log.Fatal(err)
	}
	w.Write(bs)
}

func postSession(w http.ResponseWriter, r *http.Request) {
	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	r.Body.Close()

	localSession, err := webRTC.StartClient(string(bs), width, height)
	if err != nil {
		log.Fatalln(err)
	}

	imageChannel := make(chan *image.RGBA, 2)
	go screenshotLoop(imageChannel)
	go startGame("games/supermariobros.rom", imageChannel, webRTC.InputChannel)

	w.Write([]byte(localSession))

}

func randomImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{width, height}})
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, color.RGBA{uint8(rand.Int31n(0xff)), uint8(rand.Int31n(0xff)), uint8(rand.Int31n(0xff)), 0xff - 1})
		}
	}

	return img
}

func screenshotLoop(imageChannel chan *image.RGBA) {
	for image := range imageChannel {
		if webRTC.IsConnected() {
			//rgbaImg := randomImage(width, height)
			yuv := screenshot.RgbaToYuv(image)
			webRTC.ImageChannel <- yuv
		}
	}
}
