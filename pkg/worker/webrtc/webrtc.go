package webrtc

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	vpxEncoder "github.com/poi5305/go-yuv2webRTC/vpx-encoder"
)

var config = webrtc.Configuration{
	BundlePolicy: webrtc.BundlePolicyMaxBundle,
	ICEServers: []webrtc.ICEServer{
		{URLs: []string{"stun:stun.l.google.com:19302"}},
	}}

// NewWebRTC create
func NewWebRTC() *WebRTC {
	w := &WebRTC{
		ImageChannel: make(chan []byte, 2),
		InputChannel: make(chan string, 2),
	}
	return w
}

// WebRTC connection
type WebRTC struct {
	connection  *webrtc.PeerConnection
	encoder     *vpxEncoder.VpxEncoder
	isConnected bool
	// for yuvI420 image
	ImageChannel chan []byte
	InputChannel chan string
}

// StartClient start webrtc
func (w *WebRTC) StartClient(remoteSession string, width, height int) (string, error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			w.StopClient()
		}
	}()

	// reset client
	if w.isConnected {
		w.StopClient()
		time.Sleep(2 * time.Second)
	}

	offer := webrtc.SessionDescription{}
	Decode(remoteSession, &offer)

	// safari VP8 payload default is not 96. Maybe 100
	m := webrtc.MediaEngine{}
	if err := m.PopulateFromSDP(offer); err != nil {
		panic(err)
	}
	api := webrtc.NewAPI(webrtc.WithMediaEngine(m))

	encoder, err := vpxEncoder.NewVpxEncoder(width, height, 20, 1200, 5)
	if err != nil {
		return "", err
	}
	w.encoder = encoder

	fmt.Println("=== StartClient ===")

	conn, err := api.NewPeerConnection(config)
	if err != nil {
		return "", err
	}

	conn.OnDataChannel(func(channel *webrtc.DataChannel) {
		channel.OnMessage(func(msg webrtc.DataChannelMessage) {
			//println(string(msg.Data))
			w.InputChannel <- string(msg.Data)
		})
	})

	w.connection = conn

	vp8Track, err := w.connection.NewTrack(getPayloadType(m, webrtc.RTPCodecTypeVideo, "VP8"), rand.Uint32(), "video", "robotmon")
	if err != nil {
		fmt.Println(err)
		w.StopClient()
		return "", err
	}
	_, err = w.connection.AddTrack(vp8Track)
	if err != nil {
		w.StopClient()
		return "", err
	}

	w.connection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		println(connectionState)
		if connectionState == webrtc.ICEConnectionStateConnected {
			w.isConnected = true
			w.startStreaming(vp8Track)
		}
	})

	err = w.connection.SetRemoteDescription(offer)
	if err != nil {
		fmt.Println("SetRemoteDescription error", err)
		w.StopClient()
		return "", err
	}

	answer, err := w.connection.CreateAnswer(nil)
	if err != nil {
		w.StopClient()
		return "", err
	}

	err = w.connection.SetLocalDescription(answer)
	if err != nil {
		w.StopClient()
		return "", err
	}

	gatherComplete := webrtc.GatheringCompletePromise(w.connection)

	// disable trickle ICE because we send only one full sdp
	<-gatherComplete

	answer = *w.connection.LocalDescription()
	localSession := Encode(answer)

	fmt.Println("=== StartClient Done ===")
	return localSession, nil
}

// StopClient disconnect
func (w *WebRTC) StopClient() {
	fmt.Println("===StopClient===")
	w.isConnected = false
	if w.encoder != nil {
		w.encoder.Release()
	}
	if w.connection != nil {
		w.connection.Close()
	}
	w.connection = nil
}

// IsConnected comment
func (w *WebRTC) IsConnected() bool {
	return w.isConnected
}

func (w *WebRTC) startStreaming(vp8Track *webrtc.Track) {
	fmt.Println("Start streaming")
	// send screenshot
	go func() {
		for w.isConnected {
			yuv := <-w.ImageChannel
			if len(w.encoder.Input) < cap(w.encoder.Input) {
				w.encoder.Input <- yuv
			}
		}
	}()

	// receive frame buffer
	go func() {
		for i := 0; w.isConnected; i++ {
			bs := <-w.encoder.Output
			//if i%10 == 0 {
			//	fmt.Println("On Frame", len(bs), i)
			//}
			vp8Track.WriteSample(media.Sample{Data: bs, Samples: 1})
		}
	}()
}

func Decode(in string, obj interface{}) {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(b, obj)
	if err != nil {
		panic(err)
	}
}

func Encode(obj interface{}) string {
	b, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(b)
}

func getPayloadType(m webrtc.MediaEngine, codecType webrtc.RTPCodecType, codecName string) uint8 {
	for _, codec := range m.GetCodecsByKind(codecType) {
		if codec.Name == codecName {
			return codec.PayloadType
		}
	}
	panic(fmt.Sprintf("Remote peer does not support %s", codecName))
}
