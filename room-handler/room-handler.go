package roomhandler

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

func MakeRoom(roomId string, hostId string, hostName string) Voter {
	var host = MakeVoter(hostId, hostName)
	host.IsHost = true
	var room = Room{
		Id:   roomId,
		Host: hostId,
		Voters: map[string]Voter{
			hostId: host,
		},
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
