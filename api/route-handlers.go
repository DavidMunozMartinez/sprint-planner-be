package api

import (
	"encoding/json"
	"net/http"
	utils "sprint-planner/utils"
	wshandler "sprint-planner/ws-handler"
	"time"
)

var EMPTY_SUCCESS = []byte("{ \"success\": \"true\" }")

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
	makeRoom(roomId, data.Id, data.Name)
	response.RoomId = roomId
	dispatch(w, response)
}

type JoinRoomBody struct {
	Id     string `json:"id"`
	RoomId string `json:"roomId"`
	Name   string `json:"name"`
}
type JoinRoomResponse struct {
	Room wshandler.Room `json:"room"`
}
type JoinRoomWSData struct {
	Voter wshandler.Voter `json:"voter"`
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
	var room = wshandler.GetRoomById(roomId)
	if room.Id != "" {
		var voter = makeVoter(data.Id, data.Name)

		wshandler.BroadcastToRoom(room, wshandler.WSMessage{
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
	Room wshandler.Room `json:"room"`
}

func GetRoomData(w http.ResponseWriter, r *http.Request) {
	var data GetRoomDataBody
	var response GetRoomDataResponse

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}

	response.Room = wshandler.GetRoomById(data.RoomId)
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

	var changed = wshandler.SetVote(body.RoomId, body.VoterId, body.Value)
	if changed {
		var room = wshandler.GetRoomById(body.RoomId)
		var voter = room.Voters[body.VoterId]
		wshandler.BroadcastToRoom(room, wshandler.WSMessage{
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

	var room = wshandler.GetRoomById(body.RoomId)
	room.Revealed = true
	wshandler.SetRoom(room)
	var event RevealVotesWSEvent
	event.Votes = make(map[string]float32)
	for _, v := range room.Voters {
		event.Votes[v.Id] = v.Vote
	}
	wshandler.BroadcastToRoom(room, wshandler.WSMessage{
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

	var room = wshandler.GetRoomById(body.RoomId)
	room.Revealed = false
	wshandler.SetRoom(room)

	for _, v := range room.Voters {
		wshandler.SetVote(room.Id, v.Id, v.Vote)
	}
	wshandler.BroadcastToRoom(room, wshandler.WSMessage{
		Timestamp: int(time.Now().UTC().UnixMilli()),
		Type:      "votesReset",
		// Message,
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
