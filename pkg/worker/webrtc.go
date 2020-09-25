package worker

import "github.com/pion/webrtc/v3"

func init() {
	m = webrtc.MediaEngine{}
	m.RegisterCodec(webrtc.NewRTPVP8Codec(webrtc.DefaultPayloadTypeVP8, 90000))
	m.RegisterCodec(webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000))
	api = webrtc.NewAPI(webrtc.WithMediaEngine(m))

	peerConnectionConfig = webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
	}
}

var (
	peerConnectionConfig webrtc.Configuration
	m                    webrtc.MediaEngine
	api                  *webrtc.API
)
