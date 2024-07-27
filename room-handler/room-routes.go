package roomhandler

import (
	"encoding/json"
	"net/http"
	api "sprint-planner/api"
	"sprint-planner/utils"
	wshandler "sprint-planner/ws-handler"
	"time"
)

var EMPTY_SUCCESS = []byte("{ \"success\": \"true\" }")

func Init() {
	api.SetRoute("/room-join", http.MethodPost, JoinRoom)
	api.SetRoute("/room-create", http.MethodPost, CreateRoom)
	api.SetRoute("/room-get", http.MethodGet, GetRoomData)
	api.SetRoute("/room-leave", http.MethodPost, LeaveRoom)
	api.SetRoute("/room-close", http.MethodPost, CloseRoom)
	api.SetRoute("/room-timer", http.MethodPost, StartTimer)

	api.SetRoute("/update-vote", http.MethodPost, UpdateVote)
	api.SetRoute("/reveal-votes", http.MethodPost, RevealVotes)
	api.SetRoute("/reset-votes", http.MethodPost, ResetVotes)
}

type CreateRoomBody struct {
	Id   string
	Name string
}
type CreateRoomResponse struct {
	RoomId string `json:"roomId"`
}

func CreateRoom(w http.ResponseWriter, r *http.Request) {
	var data CreateRoomBody
	var response CreateRoomResponse
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	var roomId = utils.RandStringBytes(10)
	MakeRoom(roomId, data.Id, data.Name)
	response.RoomId = roomId
	dispatch(w, response)
}

type JoinRoomBody struct {
	Id     string `json:"id"`
	RoomId string `json:"roomId"`
	Name   string `json:"name"`
}
type JoinRoomResponse struct {
	Room Room `json:"room"`
}
type JoinRoomWSData struct {
	Voter Voter `json:"voter"`
}

func JoinRoom(w http.ResponseWriter, r *http.Request) {
	var data JoinRoomBody
	var response JoinRoomResponse

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	var roomId = data.RoomId
	var room = GetRoomById(roomId)
	if room.Id != "" {
		var voter = MakeVoter(data.Id, data.Name)

		wshandler.BroadcastToRoom(room.Id, wshandler.WSMessage{
			Timestamp: int(time.Now().UTC().UnixMilli()),
			Type:      "voterJoined",
			Message: JoinRoomWSData{
				Voter: voter,
			},
			From: voter.Id,
			Room: roomId,
		})

		room.Voters[data.Id] = voter
		response.Room = room

		dispatch(w, response)
	} else {
		w.WriteHeader(http.StatusExpectationFailed)
		w.Write([]byte("ROOM_NOT_FOUND"))
	}

}

type GetRoomDataBody struct {
	RoomId string `json:"roomId"`
}
type GetRoomDataResponse struct {
	Room Room `json:"room"`
}

func GetRoomData(w http.ResponseWriter, r *http.Request) {
	var data GetRoomDataBody
	var response GetRoomDataResponse

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}

	response.Room = GetRoomById(data.RoomId)
	if response.Room.Id != "" {
		dispatch(w, response)
	} else {
		w.WriteHeader(http.StatusExpectationFailed)
		w.Write([]byte("ROOM_NOT_FOUND"))
	}
}

type UpdateVoteBody struct {
	VoterId string  `json:"voterId"`
	RoomId  string  `json:"roomId"`
	Value   float32 `json:"value"`
}
type UpdateVoteWSEvent struct {
	VoterId  string `json:"voterId"`
	Property string `json:"property"`
	Value    any    `json:"value"`
}

func UpdateVote(w http.ResponseWriter, r *http.Request) {
	var body UpdateVoteBody

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	var changed = SetVote(body.RoomId, body.VoterId, body.Value)
	if changed {
		var room = GetRoomById(body.RoomId)
		var voter = room.Voters[body.VoterId]
		wshandler.BroadcastToRoom(room.Id, wshandler.WSMessage{
			Timestamp: int(time.Now().UTC().UnixMilli()),
			Type:      "voterUpdated",
			Message: UpdateVoteWSEvent{
				VoterId:  body.VoterId,
				Property: "hasVoted",
				Value:    voter.HasVoted,
			},
			From: body.VoterId,
			Room: body.RoomId,
		})
	}
	w.WriteHeader(200)
	w.Write(EMPTY_SUCCESS)
}

type RevealVotesBoy struct {
	RoomId string `json:"roomId"`
}
type RevealVotesWSEvent struct {
	Votes map[string]float32 `json:"votes"`
}

func RevealVotes(w http.ResponseWriter, r *http.Request) {
	var body RevealVotesBoy
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	var room = GetRoomById(body.RoomId)
	room.Revealed = true
	if room.Timer.Running {
		room.Timer.Current = -1
		room.Timer.Running = false
		room.Timer.Time = -1
		var event TimerWSEvent
		event.Timer = room.Timer
		wshandler.BroadcastToRoom(body.RoomId, wshandler.WSMessage{
			Timestamp: int(time.Now().UTC().UnixMilli()),
			Type:      "timerUpdate",
			Message:   event,
		})
	}
	SetRoom(room)
	var event RevealVotesWSEvent
	event.Votes = make(map[string]float32)
	for _, v := range room.Voters {
		event.Votes[v.Id] = v.Vote
	}
	wshandler.BroadcastToRoom(room.Id, wshandler.WSMessage{
		Timestamp: int(time.Now().UTC().UnixMilli()),
		Type:      "votesRevealed",
		Message:   event,
	})
	w.WriteHeader(200)
	w.Write(EMPTY_SUCCESS)
}

type ResetVotesBody struct {
	RoomId string
}

func ResetVotes(w http.ResponseWriter, r *http.Request) {
	var body ResetVotesBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	var room = GetRoomById(body.RoomId)
	room.Revealed = false
	if room.Timer.Running {
		room.Timer.Current = -1
		room.Timer.Running = false
		room.Timer.Time = -1
		var event TimerWSEvent
		event.Timer = room.Timer
		wshandler.BroadcastToRoom(body.RoomId, wshandler.WSMessage{
			Timestamp: int(time.Now().UTC().UnixMilli()),
			Type:      "timerUpdate",
			Message:   event,
		})
	}
	SetRoom(room)

	for _, v := range room.Voters {
		SetVote(room.Id, v.Id, v.Vote)
	}
	wshandler.BroadcastToRoom(room.Id, wshandler.WSMessage{
		Timestamp: int(time.Now().UTC().UnixMilli()),
		Type:      "votesReset",
	})
	w.WriteHeader(200)
	w.Write(EMPTY_SUCCESS)
}

type LeaveRoomBody = struct {
	RoomId  string `json:"roomId"`
	VoterId string `json:"voterId"`
}
type LeaveRoomWSEvent = struct {
	VoterId string `json:"voterId"`
}

func LeaveRoom(w http.ResponseWriter, r *http.Request) {
	var body LeaveRoomBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	DeleteVoter(body.RoomId, body.VoterId)
	var event = LeaveRoomWSEvent{
		VoterId: body.VoterId,
	}
	wshandler.BroadcastToRoom(body.RoomId, wshandler.WSMessage{
		Timestamp: int(time.Now().UTC().UnixMilli()),
		Type:      "voterLeave",
		Message:   event,
	})

	w.WriteHeader(200)
	w.Write(EMPTY_SUCCESS)
}

type CloseRoomBody = struct {
	RoomId string `json:"roomId"`
}

func CloseRoom(w http.ResponseWriter, r *http.Request) {
	var body CloseRoomBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	DeleteRoom(body.RoomId)
	wshandler.BroadcastToRoom(body.RoomId, wshandler.WSMessage{
		Timestamp: int(time.Now().UTC().UnixMilli()),
		Type:      "roomClosed",
	})

	w.WriteHeader(200)
	w.Write(EMPTY_SUCCESS)
}

func dispatch(w http.ResponseWriter, response any) {
	json_data, json_error := json.Marshal(&response)
	if json_error != nil {
		w.WriteHeader(http.StatusExpectationFailed)
		w.Write([]byte(json_error.Error()))
		return
	} else {
		w.WriteHeader(200)
		w.Write(json_data)
	}
}

type StartTimerBody = struct {
	RoomId string    `json:"roomId"`
	Timer  RoomTimer `json:"timer"`
}

type TimerWSEvent struct {
	Timer RoomTimer `json:"timer"`
}

func StartTimer(w http.ResponseWriter, r *http.Request) {
	var body StartTimerBody
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	var roomTimer = body.Timer
	roomTimer.Running = true

	SetTimer(body.RoomId, roomTimer)

	// Start the async timer
	stopChan := make(chan bool)
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		stopTimer := time.NewTimer(time.Second * time.Duration(roomTimer.Time))
		defer stopTimer.Stop()
		for {
			select {
			case <-ticker.C:
				timer := TickRoomTimer(body.RoomId)
				if timer.Current < 0 {
					stopChan <- true
				}
				var event TimerWSEvent
				event.Timer = timer
				wshandler.BroadcastToRoom(body.RoomId, wshandler.WSMessage{
					Timestamp: int(time.Now().UTC().UnixMilli()),
					Type:      "timerUpdate",
					Message:   event,
				})

			case <-stopTimer.C:
				stopChan <- true
				return
			}
		}
	}()

	// Wait for the stop signal
	<-stopChan

	w.WriteHeader(200)
	w.Write(EMPTY_SUCCESS)
}
