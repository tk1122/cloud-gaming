package emulator

import (
	"context"
	"github.com/tk1122/cloud-gaming/pkg/worker/encoder"
	"image"
	"log"
	"time"

	"github.com/fogleman/nes/nes"
)

type View interface {
	Enter()
	Exit()
	Update(t, dt float64)
}

type Director struct {
	ctx          context.Context
	view         View
	timestamp    float64
	imageChannel chan *image.RGBA
	inputChannel chan string
}

func NewDirector(ctx context.Context, imageChannel chan *image.RGBA, inputChannel chan string) *Director {
	return &Director{ctx: ctx, imageChannel: imageChannel, inputChannel: inputChannel}
}

// SetView nil do saving stuff
func (d *Director) SetView(view View) {
	if d.view != nil {
		d.view.Exit()
	}
	d.view = view
	if d.view != nil {
		d.view.Enter()
	}
	d.timestamp = float64(time.Now().Nanosecond()) / float64(time.Second)
}

// Step stop being called means the game emulator stops
func (d *Director) Step() {
	timestamp := float64(time.Now().Nanosecond()) / float64(time.Second)
	dt := timestamp - d.timestamp
	d.timestamp = timestamp
	if d.view != nil {
		d.view.Update(timestamp, dt)
	}
}

func (d *Director) Start(paths []string) {
	d.PlayGame(paths[0])
	d.Run()
}

func (d *Director) Run() {
	stepTicker := time.NewTicker(time.Second / encoder.FPS)

loop:
	for range stepTicker.C {
		select {
		case <-d.ctx.Done():
			break loop
		default:
			d.Step()
		}
	}
	// closing room's image channel and input channel here
	// gurantee there is no write to closed channel from console Buffer
	// if we do closing in game room after invoking context cancel func
	// that cause write to closed channel panic due to the ticker
	close(d.imageChannel)
	close(d.inputChannel)
	d.SetView(nil)
	stepTicker.Stop()
	log.Println("Director stopped")
}

func (d *Director) PlayGame(path string) {
	hash, err := hashFile(path)
	if err != nil {
		log.Fatalln(err)
	}
	console, err := nes.NewConsole(path)
	if err != nil {
		log.Fatalln(err)
	}
	d.SetView(NewGameView(d, console, hash, d.imageChannel, d.inputChannel))
}
