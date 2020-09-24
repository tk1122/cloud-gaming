package worker

import (
	"github.com/gorilla/mux"
	"github.com/pion/webrtc/v3"
	"github.com/poi5305/go-yuv2webRTC/screenshot"
	"github.com/tk1122/cloud-gaming/pkg/worker/emulator"
	"github.com/tk1122/cloud-gaming/pkg/worker/encoder"
	"image"
)

var (
	Router *mux.Router
	e      *encoder.Encoder
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func init() {
	genPem()
}

func init() {
	Router = mux.NewRouter()
	Router.HandleFunc("/", getWeb).Methods("GET")
	Router.HandleFunc("/ws", getWs).Methods("GET")
}

func startSession(track *webrtc.Track) {
	e = encoder.NewEncoder()
	imageChannel := make(chan *image.RGBA, 2)

	go startGame("games/supermariobros.rom", imageChannel, e.InputChannel)
	go screenshotLoop(e, imageChannel)
	e.StartStreaming(track)
}

func sendInputToSession(input string) {
	e.InputChannel <- input
}

func stopSession() {
	if e != nil {
		e.StopStreaming()
	}
}

func screenshotLoop(e *encoder.Encoder, imageChannel chan *image.RGBA) {
	for i := range imageChannel {
		if e.IsRunning() {
			yuv := screenshot.RgbaToYuv(i)
			e.ImageChannel <- yuv
		}
	}
}

func startGame(path string, imageChannel chan *image.RGBA, inputChannel chan string) {
	director := emulator.NewDirector(imageChannel, inputChannel)
	director.Start([]string{path})
}
