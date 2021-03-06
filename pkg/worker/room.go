package worker

import (
	"context"
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
	clientsMux     sync.Mutex
	clients        map[int]*client
	id             string
	encoder        *encoder.Encoder
	director       *emulator.Director
	cancelDirector context.CancelFunc
	imageChanel    chan *image.RGBA
	inputChanel    chan string
	isRunning      bool
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
	r.clientsMux.Lock()
	defer r.clientsMux.Unlock()

	r.clients[int(c.playerId)] = c
	c.room = r
	r.encoder.AddTrack(c.track, int(c.playerId))
}

func (r *room) removeClient(c *client) {
	r.clientsMux.Lock()
	defer r.clientsMux.Unlock()

	r.encoder.RemoveTrack(c.track, int(c.playerId))
	c.room = nil
	delete(r.clients, int(c.playerId))
}

func (r *room) receiveInputMessage(input string) {
	r.inputChanel <- input
}

func (r *room) joinOrStartGame(gameId string) {
	log.Println("Player join game")
	if r.isRunning {
		log.Println("Game is already running")
		return
	}
	log.Println("Start new game")
	r.isRunning = true

	ctx, ctxCancle := context.WithCancel(context.Background())
	r.cancelDirector = ctxCancle

	go r.startGame(ctx, "games/"+gameId)
	go r.screenshotLoop()
	r.encoder.StartStreaming()
}

func (r *room) leaveOrStopGame(c *client) {
	log.Println("Player leave game")
	r.removeClient(c)

	if len(r.clients) == 0 && r.isRunning {
		r.isRunning = false
		r.cancelDirector()
		r.encoder.StopStreaming()

		delete(rooms, r.id)
		log.Println("Game stopped")
	}
}

func (r *room) screenshotLoop() {
	for i := range r.imageChanel {
		if r.encoder.IsRunning() {
			yuv := screenshot.RgbaToYuv(i)
			r.encoder.ImageChannel <- yuv
		}
	}
}

func (r *room) startGame(ctx context.Context, path string) {
	r.director = emulator.NewDirector(ctx, r.imageChanel, r.inputChanel)
	r.director.Start([]string{path})
}
