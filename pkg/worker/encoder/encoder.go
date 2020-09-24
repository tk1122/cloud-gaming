package encoder

import (
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	vpxEncoder "github.com/poi5305/go-yuv2webRTC/vpx-encoder"
	"log"
)

const (
	width  = 256
	height = 240
)

type Encoder struct {
	encoder      *vpxEncoder.VpxEncoder
	isRunning    bool
	ImageChannel chan []byte
	InputChannel chan string
}

func NewEncoder() *Encoder {
	encoder, err := vpxEncoder.NewVpxEncoder(width, height, 20, 1200, 5)
	must(err)

	return &Encoder{
		encoder:      encoder,
		isRunning:    false,
		ImageChannel: make(chan []byte, 2),
		InputChannel: make(chan string, 2),
	}
}

func (e *Encoder) IsRunning() bool {
	return e.isRunning
}

func (e *Encoder) StopStreaming() {
	e.isRunning = false
	e.encoder.Release()
}

func (e *Encoder) StartStreaming(vp8Track *webrtc.Track) {
	log.Println("Start streaming")
	e.isRunning = true

	go func() {
		for e.isRunning {
			yuv := <-e.ImageChannel
			if len(e.encoder.Input) < cap(e.encoder.Input) {
				e.encoder.Input <- yuv
			}
		}
	}()

	go func() {
		for e.isRunning {
			bs := <-e.encoder.Output
			_ = vp8Track.WriteSample(media.Sample{Data: bs, Samples: 1})
		}
	}()
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
