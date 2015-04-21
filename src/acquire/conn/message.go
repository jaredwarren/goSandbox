package conn

import (
	"acquire/game"
)

// General message struct which is used for parsing client requests and sending
// back responses.
type Message struct {
	Action     string `json:"action"`
	Turn       int    `json:"turn"`
	NumPlayers int32
	History    string
	Text       string
}
