package emulator

import (
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
	view         View
	menuView     View
	timestamp    float64
	imageChannel chan *image.RGBA
	inputChannel chan string
}

func NewDirector(imageChannel chan *image.RGBA, inputChannel chan string) *Director {
	return &Director{imageChannel: imageChannel, inputChannel: inputChannel}
}

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
	for {
		d.Step()
		time.Sleep(time.Second / encoder.FPS)
	}
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
	d.SetView(NewGameView(d, console, path, hash, d.imageChannel, d.inputChannel))
}
