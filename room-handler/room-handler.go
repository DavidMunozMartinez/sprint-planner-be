package roomhandler

import (
	wshandler "sprint-planner/ws-handler"
	"time"
)

var rooms = make(map[string]Room)

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

func SetTimer(roomId string, timer RoomTimer) {
	var room = rooms[roomId]
	room.Timer = timer
	rooms[roomId] = room
}

func TickRoomTimer(roomId string) RoomTimer {
	var room = rooms[roomId]
	room.Timer.Current = room.Timer.Current - 1
	if room.Timer.Current == 0 {
		room.Timer.Running = false
	}
	rooms[roomId] = room
	return room.Timer
}

func MakeRoom(roomId string, hostId string, hostName string) Voter {
	var host = MakeVoter(hostId, hostName)
	host.IsHost = true
	var room = Room{
		Id:   roomId,
		Host: hostId,
		Voters: map[string]Voter{
			hostId: host,
		},
		LastActivity: time.Now(),
	}
	SetRoom(room)
	return host
}

func MakeVoter(voterId string, voterName string) Voter {
	return Voter{
		Id:     voterId,
		Name:   voterName,
		Vote:   -1,
		IsHost: false,
	}
}

func DeleteVoter(roomId string, voterId string) {
	delete(rooms[roomId].Voters, voterId)
}

func DeleteRoom(roomId string) {
	delete(rooms, roomId)
}

func StartInactiveRoomChecker() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				now := time.Now()
				for roomId, room := range rooms {
					if now.Sub(room.LastActivity) > 1*time.Minute {
						// Room has been inactive for more than 10 minutes
						CloseRoomById(roomId)
					}
				}
			}
		}
	}()
}

func CloseRoomById(roomId string) {
	DeleteRoom(roomId)
	wshandler.BroadcastToRoom(roomId, wshandler.WSMessage{
		Timestamp: int(time.Now().UTC().UnixMilli()),
		Type:      "roomClosed",
	})
}
