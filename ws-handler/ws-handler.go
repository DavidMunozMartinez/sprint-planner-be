package wshandler

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type WSMessage struct {
	Timestamp int    `json:"timestamp"`
	Type      string `json:"type"`
	Message   any    `json:"message"`
	From      string `json:"from"`
	Room      string `json:"room"`
}

type Room struct {
	Id       string           `json:"id"`
	Host     string           `json:"host"`
	Voters   map[string]Voter `json:"voters"`
	Revealed bool             `json:"revealed"`
}

type Voter struct {
	Id       string          `json:"id"`
	Name     string          `json:"name"`
	Vote     float32         `json:"vote"`
	HasVoted bool            `json:"hasVoted"`
	IsHost   bool            `json:"isHost"`
	WS       *websocket.Conn `json:"-"`
}

var rooms = make(map[string]Room)
var broadcast = make(chan WSMessage)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Init() {
	http.HandleFunc("/ws", handleConnections)
}

func GetRoomById(roomId string) Room {
	return rooms[roomId]
}

func SetRoom(room Room) {
	rooms[room.Id] = room
}

func SetVote(roomId string, voterId string, vote float32) bool {
	var room = rooms[roomId]
	var voter = room.Voters[voterId]
	var currentHasVotedState = voter.HasVoted
	voter.HasVoted = voter.Vote != vote
	var hasVotedChanged = currentHasVotedState != voter.HasVoted
	if voter.HasVoted {
		voter.Vote = vote
	} else {
		voter.Vote = -1
	}
	rooms[roomId].Voters[voterId] = voter
	return hasVotedChanged
}

func SetWS(roomId string, voterId string, ws *websocket.Conn) {
	var room = rooms[roomId]
	var voter = room.Voters[voterId]
	voter.WS = ws
	rooms[roomId].Voters[voterId] = voter
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	var roomId = query.Get("room")
	var id = query.Get("id")
	SetWS(roomId, id, ws)

	for {
		var msg WSMessage
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(rooms, query.Get("id"))
			break
		}
		broadcast <- msg
	}
}

func BroadcastToRoom(room Room, msg WSMessage) {
	for i, v := range room.Voters {
		if v.WS != nil {
			err := v.WS.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				v.WS.Close()
				delete(room.Voters, i)
			}
		}
	}
}
