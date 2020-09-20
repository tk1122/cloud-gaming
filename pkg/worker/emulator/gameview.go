package emulator

import (
	"image"

	"github.com/fogleman/nes/nes"
)

const padding = 0

type GameView struct {
	director   *Director
	console    *nes.Console
	title      string
	hash       string
	record     bool
	frames     []image.Image
	keyPressed [8]bool

	imageChannel chan *image.RGBA
	inputChannel chan string
}

func NewGameView(director *Director, console *nes.Console, title, hash string, imageChannel chan *image.RGBA, inputChannel chan string) View {
	gameview := &GameView{director, console, title, hash, false, nil, [8]bool{false}, imageChannel, inputChannel}
	go gameview.listenToInputChannel()

	return gameview
}

func (view *GameView) listenToInputChannel() {
	for keyString := range view.inputChannel {
		println(keyString)
		for id, c := range keyString {
			if c == '1' {
				view.keyPressed[id] = true
			} else {
				view.keyPressed[id] = false
			}
		}
	}
}

func (view *GameView) updateControllers() {
	view.console.Controller1.SetButtons(view.keyPressed)
}

func (view *GameView) Enter() {
	// load state
	if err := view.console.LoadState(savePath(view.hash)); err == nil {
		return
	} else {
		view.console.Reset()
	}
	// load sram
	cartridge := view.console.Cartridge
	if cartridge.Battery != 0 {
		if sram, err := readSRAM(sramPath(view.hash)); err == nil {
			cartridge.SRAM = sram
		}
	}
}

func (view *GameView) Exit() {
	// save sram
	cartridge := view.console.Cartridge
	if cartridge.Battery != 0 {
		writeSRAM(sramPath(view.hash), cartridge.SRAM)
	}
	// save state
	view.console.SaveState(savePath(view.hash))
}

func (view *GameView) Update(t, dt float64) {
	if dt > 1 {
		dt = 0
	}
	console := view.console
	view.updateControllers()
	console.StepSeconds(dt)
	view.imageChannel <- console.Buffer()
	if view.record {
		view.frames = append(view.frames, copyImage(console.Buffer()))
	}
}
