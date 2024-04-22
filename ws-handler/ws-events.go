package wshandler

type VoterJoinedEvent struct {
	Timestamp int   `json:"timestamp"`
	Voter     Voter `json:"voter"`
}

type VoterUpdatedEvent struct {
	Timestamp int    `json:"timestamp"`
	Property  string `json:"property"`
	Value     any    `json:"value"`
}
