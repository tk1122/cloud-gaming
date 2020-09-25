package worker

import (
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

type client struct {
	wsConn   *websocket.Conn
	peerConn *webrtc.PeerConnection
	track    *webrtc.Track
	playerId int
	room     *room
}

func newClient(wsConn *websocket.Conn, peerConn *webrtc.PeerConnection, track *webrtc.Track, playerId int) *client {
	return &client{
		wsConn:   wsConn,
		peerConn: peerConn,
		track:    track,
		playerId: playerId,
	}
}

func (c *client) joinGame() {
	c.room.joinOrStartGame()
}

func (c *client) leaveGame() {
	c.room.leaveOrStopGame()
}

func (c *client) sendInputToGame(input string) {
	c.room.receiveInputMessage(input)
}
