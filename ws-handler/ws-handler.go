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

var wsrooms = make(map[string]map[string]*websocket.Conn)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Init() {
	http.HandleFunc("/ws", handleConnections)
}

func SetWS(roomId string, userId string, ws *websocket.Conn) {
	if wsrooms[roomId] == nil {
		wsrooms[roomId] = make(map[string]*websocket.Conn)
	}
	wsrooms[roomId][userId] = ws
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		ws.Close()
	}

	var roomId = query.Get("room")
	var id = query.Get("id")
	SetWS(roomId, id, ws)
}

func BroadcastToRoom(roomId string, msg WSMessage) {
	for i, v := range wsrooms[roomId] {
		err := v.WriteJSON(msg)
		if err != nil {
			log.Printf("error: %v", err)
			v.Close()
			delete(wsrooms[roomId], i)
		}
	}
}
