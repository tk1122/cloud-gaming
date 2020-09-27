package worker

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"log"
	"strconv"
	"time"
)

type client struct {
	wsConn   *websocket.Conn
	peerConn *webrtc.PeerConnection
	track    *webrtc.Track
	playerId playerId
	room     *room
}

type playerId int

const (
	PlayerOne playerId = iota + 1
	PlayerTwo
)

func newClient(wsConn *websocket.Conn, peerConn *webrtc.PeerConnection, track *webrtc.Track) *client {
	return &client{
		wsConn:   wsConn,
		peerConn: peerConn,
		track:    track,
	}
}

func (client *client) joinOrStartGame() {
	client.room.joinOrStartGame()
}

func (client *client) leaveOrStopGame() {
	client.room.leaveOrStopGame()
}

func (client *client) sendInputToGame(input string) {

	// TODO try another fancier way
	if client.playerId == PlayerTwo {
		playerOneKeysPlaceholder := "0000000000"
		input = playerOneKeysPlaceholder + input
	}

	client.room.receiveInputMessage(input)
}

func (client *client) setPlayerId(id playerId) {
	client.playerId = id
}

func (client *client) registerICEConnectionEvents(pendingCandidates []*webrtc.ICECandidate) {
	client.peerConn.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Println("ice connection state changed:", state)
		switch state {
		case webrtc.ICEConnectionStateConnected:
			client.joinOrStartGame()
		case webrtc.ICEConnectionStateClosed:
			client.room.removeClient(client)
			client.leaveOrStopGame()
		case webrtc.ICEConnectionStateFailed:
			_ = client.wsConn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
				time.Now().Add(writeWait),
			)
			time.Sleep(closeGracePeriod)
			_ = client.wsConn.Close()
			_ = client.peerConn.Close()
			log.Println("player left: websocket and peer connections closed")
		}
	})

	client.peerConn.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		if remoteDesc := client.peerConn.RemoteDescription(); remoteDesc != nil {
			iceCandidate, err := json.Marshal(c.ToJSON())
			must(err)

			// candidate ws message sent to browser does not contain roomId and playerId
			packet := newWsPacket(Candidate, string(iceCandidate), "", 0)
			sendMessage(client.wsConn, websocket.TextMessage, packet)
		} else {
			log.Println("get new ice candidate but remote description yet set")
			pendingCandidates = append(pendingCandidates, c)
		}
	})

	client.peerConn.OnDataChannel(func(channel *webrtc.DataChannel) {
		channel.OnMessage(func(msg webrtc.DataChannelMessage) {
			// ignore input when game is not running
			if !client.room.isRunning {
				return
			}

			client.sendInputToGame(string(msg.Data))
		})
	})
}

func (client *client) listenPeerMessages(pendingCandidate []*webrtc.ICECandidate) {
	defer func() {
		log.Println("process ws message failed: websocket connection closed")
		_ = client.wsConn.Close()
	}()

	//room for entire websocket connection
	//assume there is only one game available
	var room *room

	client.wsConn.SetReadLimit(maxMessageSize)
	must(client.wsConn.SetReadDeadline(time.Now().Add(pongWait)))

	// forever loop to listening incoming messages
	for {
		mt, msg, err := client.wsConn.ReadMessage()
		if err != nil {
			break
		}

		req := &wsPacket{}
		err = json.Unmarshal(msg, req)
		if err != nil {
			break
		}

		switch req.Id {
		case Offer:
			log.Println("received offer")
			if req.PlayerId != int(PlayerOne) && req.PlayerId != int(PlayerTwo) {
				log.Println("only support two player")
				break
			}

			if req.PlayerId == int(PlayerOne) {
				client.setPlayerId(PlayerOne)

				// create new room in case of player 1
				room = newRoom()
				room.addClient(client)
			}

			if req.PlayerId == int(PlayerTwo) {
				ok := false
				if room, ok = rooms[req.RoomId]; !ok {
					log.Println("room not found")
					break
				}
				client.setPlayerId(PlayerTwo)
				room.addClient(client)
			}

			log.SetPrefix("room id:" + room.id + " player:" + strconv.Itoa(int(client.playerId)) + " ")

			err = client.peerConn.SetRemoteDescription(webrtc.SessionDescription{
				SDP:  req.Data,
				Type: webrtc.SDPTypeOffer,
			})

			if err != nil {
				log.Println("cannot set remote sdp: ", err)
				break
			}

			answer, err := client.peerConn.CreateAnswer(nil)
			must(err)
			must(client.peerConn.SetLocalDescription(answer))

			packet := newWsPacket(Answer, answer.SDP, room.id, client.playerId)
			sendMessage(client.wsConn, mt, packet)
			log.Println("answer sent")

			for _, cdd := range pendingCandidate {
				// candidate ws message sent to browser does not contain roomId and playerId
				packet := newWsPacket(Candidate, cdd.ToJSON().Candidate, "", 0)
				sendMessage(client.wsConn, mt, packet)
			}
			log.Println("all pending candidate sent")
		case Candidate:
			log.Println("received new remote ice candidate")

			err = client.peerConn.AddICECandidate(webrtc.ICECandidateInit{
				Candidate: req.Data,
			})
			if err != nil {
				log.Println("cannot add remote ice candidate: ", err)
				break
			}
		default:
			break
		}
	}
}
