package roomhandler

import "time"

type RoomTimer struct {
	Running bool `json:"running"`
	Time    int  `json:"time"`
	Current int  `json:"current"`
}

type Room struct {
	Id           string           `json:"id"`
	Host         string           `json:"host"`
	Voters       map[string]Voter `json:"voters"`
	Revealed     bool             `json:"revealed"`
	Timer        RoomTimer        `json:"timer"`
	LastActivity time.Time        `json:"-"`
}

type Voter struct {
	Id       string  `json:"id"`
	RoomId   string  `json:"roomId"`
	Name     string  `json:"name"`
	Vote     float32 `json:"vote"`
	HasVoted bool    `json:"hasVoted"`
	IsHost   bool    `json:"isHost"`
}
