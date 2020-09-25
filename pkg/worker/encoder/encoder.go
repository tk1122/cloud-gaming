package encoder

import (
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	vpxEncoder "github.com/poi5305/go-yuv2webRTC/vpx-encoder"
	"log"
)

type Encoder struct {
	encoder      *vpxEncoder.VpxEncoder
	isRunning    bool
	ImageChannel chan []byte
	tracks       map[int]*webrtc.Track
}

const (
	FPS    = 60
	WIDTH  = 256
	HEIGHT = 240
)

func NewEncoder() *Encoder {
	encoder, err := vpxEncoder.NewVpxEncoder(WIDTH, HEIGHT, FPS, 1200, 5)
	must(err)

	return &Encoder{
		encoder:      encoder,
		tracks:       make(map[int]*webrtc.Track),
		isRunning:    false,
		ImageChannel: make(chan []byte, 5),
	}
}

func (e *Encoder) IsRunning() bool {
	return e.isRunning
}

func (e *Encoder) StopStreaming() {
	e.isRunning = false
	e.encoder.Release()
}

func (e *Encoder) AddTrack(t *webrtc.Track, playerId int) {
	e.tracks[playerId] = t
}

func (e *Encoder) RemoveTrack(t *webrtc.Track, playerId int) {
	delete(e.tracks, playerId)
}

func (e *Encoder) StartStreaming() {
	if e.isRunning {
		log.Println("Already start streaming")
		return
	}

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
			// encoded once, send to multiple webrtc tracks
			bs := <-e.encoder.Output
			for _, t := range e.tracks {
				go func(t *webrtc.Track) {
					_ = t.WriteSample(media.Sample{Data: bs, Samples: 1})
				}(t)
			}
		}
	}()
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
