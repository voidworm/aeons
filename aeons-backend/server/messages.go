package server

import "encoding/json"

const (
	MsgNewGame = "new_game"
	MsgGameState = "game_state"
	MsgError = "error"
)

type Envelope struct {
	Type string `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}