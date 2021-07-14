package chess

// this file contains the data structures used for creating and holding
// the unmarshalled json data returned from the requests of the main
// chess.go file.

const (
	leaderboards = "leaderboards"
)

type ResponseLeaderboards struct {
	PlayerId int    `json:"player_id"`
	Username string `json:"username"`
	Rank     int    `json:"rank"`
	Score    int    `json:"score"`
}

type Leaderboards struct {
	Daily []ResponseLeaderboards `json:"daily"`
}
