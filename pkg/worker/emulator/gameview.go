package emulator

import (
	"github.com/fogleman/nes/nes"
	"image"
)

type GameView struct {
	director     *Director
	console      *nes.Console
	hash         string
	keyPressed   [playerKeyNums * 2]bool
	imageChannel chan *image.RGBA
	inputChannel chan string
}

const (
	PlayerOneFirstBit = "0"
	PlayerTwoFirstBit = "1"
	playerKeyNums     = 8
)

func NewGameView(director *Director, console *nes.Console, hash string, imageChannel chan *image.RGBA, inputChannel chan string) View {
	gameview := &GameView{
		console:      console,
		director:     director,
		hash:         hash,
		inputChannel: inputChannel,
		imageChannel: imageChannel,
		keyPressed:   [playerKeyNums * 2]bool{false},
	}
	go gameview.listenToInputChannel()

	return gameview
}

func (view *GameView) listenToInputChannel() {
	for keyString := range view.inputChannel {
		bitOffset := 0
		if string(keyString[0]) == PlayerTwoFirstBit {
			bitOffset = playerKeyNums
		}
		for id, c := range keyString[1:] {
			if c == '1' {
				view.keyPressed[id+bitOffset] = true
			} else {
				view.keyPressed[id+bitOffset] = false
			}
		}
	}
}

func (view *GameView) updateControllers() {
	var player1Keys, player2Keys [playerKeyNums]bool
	copy(player1Keys[:], view.keyPressed[:playerKeyNums])
	copy(player2Keys[:], view.keyPressed[playerKeyNums:])
	view.console.Controller1.SetButtons(player1Keys)
	view.console.Controller2.SetButtons(player2Keys)
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
}
