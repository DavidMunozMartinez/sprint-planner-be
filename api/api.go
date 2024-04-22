package api

import (
	"net/http"
	wshandler "sprint-planner/ws-handler"
)

type Route = struct {
	path    string
	handler func(w http.ResponseWriter, r *http.Request)
}

var Routes = []Route{
	{
		path:    "/room-join",
		handler: JoinRoom,
	},
	{
		path:    "/room-create",
		handler: CreateRoom,
	},
	{
		path:    "/room-get",
		handler: GetRoomData,
	},
	{
		path:    "/update-vote",
		handler: UpdateVote,
	},
	{
		path:    "/reveal-votes",
		handler: RevealVotes,
	},
	{
		path:    "/reset-votes",
		handler: ResetVotes,
	},
}

func InitRoutes() {
	for _, route := range Routes {
		http.HandleFunc(route.path, func(w http.ResponseWriter, r *http.Request) {
			if validator(w, r) {
				route.handler(w, r)
			} else {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte((`"error": "bad call"`)))
			}
		})
	}
}

// Runs before each route
func validator(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	return true
}

func makeRoom(roomId string, hostId string, hostName string) wshandler.Voter {
	var host = makeVoter(hostId, hostName)
	host.IsHost = true
	var room = wshandler.Room{
		Id:   roomId,
		Host: hostId,
		Voters: map[string]wshandler.Voter{
			hostId: host,
		},
	}
	wshandler.SetRoom(room)
	return host
}

func makeVoter(voterId string, voterName string) wshandler.Voter {
	return wshandler.Voter{
		Id:     voterId,
		Name:   voterName,
		Vote:   -1,
		IsHost: false,
	}
}
