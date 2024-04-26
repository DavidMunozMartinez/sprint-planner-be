package roomhandler

type Room struct {
	Id       string           `json:"id"`
	Host     string           `json:"host"`
	Voters   map[string]Voter `json:"voters"`
	Revealed bool             `json:"revealed"`
}

type Voter struct {
	Id       string  `json:"id"`
	RoomId   string  `json:"roomId"`
	Name     string  `json:"name"`
	Vote     float32 `json:"vote"`
	HasVoted bool    `json:"hasVoted"`
	IsHost   bool    `json:"isHost"`
}
