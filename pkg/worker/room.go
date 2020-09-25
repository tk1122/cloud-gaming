package worker

import (
	"github.com/poi5305/go-yuv2webRTC/screenshot"
	"github.com/tk1122/cloud-gaming/pkg/worker/emulator"
	"github.com/tk1122/cloud-gaming/pkg/worker/encoder"
	"image"
	"log"
	"math/rand"
	"strconv"
	"sync"
)

type room struct {
	clients     map[int]*client
	encoder     *encoder.Encoder
	imageChanel chan *image.RGBA
	inputChanel chan string
	isRunning   bool
}

var rooms = make(map[string]*room)

func generateRoomID() string {
	roomID := strconv.FormatInt(rand.Int63(), 16)
	return roomID
}

func newRoom() *room {
	r := &room{
		clients:     make(map[int]*client),
		encoder:     encoder.NewEncoder(),
		imageChanel: make(chan *image.RGBA, 5),
		inputChanel: make(chan string),
		isRunning:   false,
	}

	roomId := generateRoomID()
	rooms[roomId] = r

	return r
}

func (r *room) addClient(c *client) {
	var mux sync.Mutex
	mux.Lock()
	defer mux.Unlock()

	r.clients[c.playerId] = c
	r.encoder.AddTrack(c.track, c.playerId)
}

func (r *room) removeClient(c *client) {
	var mux sync.Mutex
	mux.Lock()
	defer mux.Unlock()

	r.encoder.RemoveTrack(c.track, c.playerId)
	delete(r.clients, c.playerId)
}

// TODO still send input in 1-player style
func (r *room) receiveInputMessage(input string) {
	r.inputChanel <- input
}

func (r *room) joinOrStartGame() {
	if !r.isRunning {
		log.Println("Start new game")
		r.isRunning = true

		go startGame("games/supermariobros.rom", r.imageChanel, r.inputChanel)
		go screenshotLoop(r.encoder, r.imageChanel)
		r.encoder.StartStreaming()
	}
}

// TODO implement
func (r *room) leaveOrStopGame() {

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
