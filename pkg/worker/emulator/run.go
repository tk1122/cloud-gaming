package emulator

import (
	"image"
	"runtime"
)

const (
	width  = 256
	height = 240
	scale  = 3
	title  = "NES"
)

func init() {
	// we need a parallel OS thread to avoid audio stuttering
	runtime.GOMAXPROCS(2)

	// we need to keep OpenGL calls on a single thread
	runtime.LockOSThread()
}

func Run(paths []string, imageChannel chan *image.RGBA, inputChannel chan string) {
	director := NewDirector(imageChannel, inputChannel)
	director.Start(paths)
}
