package worker

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type wsPacketID string
type wsPacket struct {
	ID   wsPacketID `json:"ID"`
	Data string     `json:"Data"`
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 8192

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Time to wait before force close on connection.
	closeGracePeriod = 10 * time.Second
)

const (
	Offer     wsPacketID = "offer"
	Answer               = "answer"
	Candidate            = "candidate"
)

var (
	upgrader = websocket.Upgrader{}
	m        webrtc.MediaEngine
	api      *webrtc.API
)

func getWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	must(err)

	ws.SetPongHandler(func(appData string) error {
		must(ws.SetReadDeadline(time.Now().Add(pongWait)))
		return nil
	})

	pc, err := api.NewPeerConnection(peerConnectionConfig)
	must(err)

	vp8Track, err := pc.NewTrack(
		webrtc.DefaultPayloadTypeVP8,
		rand.Uint32(),
		fmt.Sprintf("video-%d", rand.Uint32()),
		fmt.Sprintf("video-%d", rand.Uint32()),
	)
	must(err)
	_, err = pc.AddTrack(vp8Track)
	must(err)

	pendingCandidates := make([]*webrtc.ICECandidate, 0)

	go ping(ws)
	go listenPeerMessages(ws, pc, pendingCandidates)

	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Println("Ice connection state changed: ", state)
		switch state {
		case webrtc.ICEConnectionStateConnected:
			startSession(vp8Track)
		case webrtc.ICEConnectionStateDisconnected, webrtc.ICEConnectionStateClosed, webrtc.ICEConnectionStateFailed:
			stopSession()
		}
	})

	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		if remoteDesc := pc.RemoteDescription(); remoteDesc != nil {
			iceCandidate, err := json.Marshal(c.ToJSON())
			must(err)

			sendMessage(ws, websocket.TextMessage, Candidate, string(iceCandidate))
		} else {
			pendingCandidates = append(pendingCandidates, c)
		}
	})

	pc.OnDataChannel(func(channel *webrtc.DataChannel) {
		channel.OnMessage(func(msg webrtc.DataChannelMessage) {
			sendInputToSession(string(msg.Data))
		})
	})
}

func ping(ws *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = ws.Close()
	}()

	for {
		<-ticker.C
		if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Time{}); err != nil {
			log.Println("ping:", err)
			break
		}
	}
}

func listenPeerMessages(ws *websocket.Conn, pc *webrtc.PeerConnection, pendingCandidate []*webrtc.ICECandidate) {
	defer func() {
		_ = ws.Close()
	}()

	// forever loop to listening incomming messages
	for {
		mt, msg, err := ws.ReadMessage()
		if err != nil {
			break
		}

		req := &wsPacket{}
		must(json.Unmarshal(msg, req))

		switch req.ID {
		case Offer:
			log.Println("Received offer")

			must(pc.SetRemoteDescription(webrtc.SessionDescription{
				SDP:  req.Data,
				Type: webrtc.SDPTypeOffer,
			}))

			answer, err := pc.CreateAnswer(nil)
			must(err)
			must(pc.SetLocalDescription(answer))

			sendMessage(ws, mt, Answer, answer.SDP)
			log.Println("Answer sended")

			for _, c := range pendingCandidate {
				sendMessage(ws, mt, Candidate, c.ToJSON().Candidate)
			}
			log.Println("All pending candidate sended")
		case Candidate:
			log.Println("Received new ice candidate")
			must(pc.AddICECandidate(webrtc.ICECandidateInit{
				Candidate: req.Data,
			}))
		}
	}
}

func sendMessage(ws *websocket.Conn, messageType int, id wsPacketID, data string) {
	var sendMux sync.Mutex
	res := &wsPacket{
		ID:   id,
		Data: data,
	}

	resMsg, err := json.Marshal(res)
	must(err)

	sendMux.Lock()
	defer sendMux.Unlock()
	if err := ws.WriteMessage(messageType, resMsg); err != nil {
		_ = ws.Close()
	}
}