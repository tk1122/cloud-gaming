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
	ID   wsPacketID `json:"id"`
	Data string     `json:"data"`
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
)

func getWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	must(err)

	ws.SetReadLimit(maxMessageSize)
	_ = ws.SetReadDeadline(time.Now().Add(pongWait))
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

	// create both new client and new room
	// TODO remove hard coded playerId and force create room
	c := newClient(ws, pc, vp8Track, 1)
	room := newRoom()
	room.addClient(c)
	c.room = room

	pendingCandidates := make([]*webrtc.ICECandidate, 0)

	go ping(ws)
	go listenPeerMessages(ws, pc, pendingCandidates)

	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Println("Ice connection state changed: ", state)
		switch state {
		case webrtc.ICEConnectionStateConnected:
			c.joinGame()
		case webrtc.ICEConnectionStateClosed, webrtc.ICEConnectionStateFailed:
			c.leaveGame()

			_ = ws.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
				time.Now().Add(writeWait),
			)
			time.Sleep(closeGracePeriod)
			_ = ws.Close()
			log.Println("Websocket connection closed")
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
			c.sendInputToGame(string(msg.Data))
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
		if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
			log.Println("ping:", err)
			break
		}
	}
}

func listenPeerMessages(ws *websocket.Conn, pc *webrtc.PeerConnection, pendingCandidate []*webrtc.ICECandidate) {
	defer func() {
		_ = ws.Close()
	}()

	// forever loop to listening incoming messages
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
			log.Println("Answer sent")

			for _, c := range pendingCandidate {
				sendMessage(ws, mt, Candidate, c.ToJSON().Candidate)
			}
			log.Println("All pending candidate sent")
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

	// to avoid concurrent write to one ws connection
	sendMux.Lock()
	defer sendMux.Unlock()

	_ = ws.SetWriteDeadline(time.Now().Add(writeWait))
	if err := ws.WriteMessage(messageType, resMsg); err != nil {
		_ = ws.Close()
	}
}

func closeConn(ws *websocket.Conn) {
	_ = ws.Close()
}
