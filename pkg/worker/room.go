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
	id          string
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
	r.id = roomId
	rooms[roomId] = r

	return r
}

func (r *room) addClient(c *client) {
	var mux sync.Mutex
	mux.Lock()
	defer mux.Unlock()

	r.clients[int(c.playerId)] = c
	c.room = r
	r.encoder.AddTrack(c.track, int(c.playerId))
}

func (r *room) removeClient(c *client) {
	var mux sync.Mutex
	mux.Lock()
	defer mux.Unlock()

	r.encoder.RemoveTrack(c.track, int(c.playerId))
	c.room = nil
	delete(r.clients, int(c.playerId))
}

// TODO still send input in 1-player style
func (r *room) receiveInputMessage(input string) {
	r.inputChanel <- input
}

func (r *room) joinOrStartGame() {
	if !r.isRunning {
		log.Println("Start new game")
		r.isRunning = true

		go r.startGame("games/contra.rom")
		go r.screenshotLoop()
		r.encoder.StartStreaming()
	}
}

// TODO implement
func (r *room) leaveOrStopGame() {

}

func (r *room) screenshotLoop() {
	for i := range r.imageChanel {
		if r.encoder.IsRunning() {
			yuv := screenshot.RgbaToYuv(i)
			r.encoder.ImageChannel <- yuv
		}
	}
}

func (r *room) startGame(path string) {
	director := emulator.NewDirector(r.imageChanel, r.inputChanel)
	director.Start([]string{path})
}
