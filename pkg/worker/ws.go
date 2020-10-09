package worker

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type wsPacketID string
type wsPacket struct {
	Id       wsPacketID `json:"id"`
	Data     string     `json:"data"`
	RoomId   string     `json:"room_id"`
	PlayerId int        `json:"player_id"`
	GameId   string     `json:"game_id"`
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 16 * 1024

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
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
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

	// client only join room and assigned playerId after sending webrtc offer
	client := newClient(ws, pc, vp8Track)
	pendingCandidates := make([]*webrtc.ICECandidate, 0)

	go ping(client.wsConn)
	go client.listenPeerMessages(pendingCandidates)
	client.registerICEConnectionEvents(pendingCandidates)
}

func ping(ws *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		log.Println("ping failed: websocket connection closed")
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

func newWsPacket(id wsPacketID, data string, roomId string, playerId playerId) *wsPacket {
	return &wsPacket{
		Id:       id,
		Data:     data,
		RoomId:   roomId,
		PlayerId: int(playerId),
	}
}
