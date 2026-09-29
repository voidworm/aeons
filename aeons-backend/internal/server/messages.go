package server

import "encoding/json"

const (
	// client to server messages
	MsgNewGame = "new_game"
	MsgSelectPlayer = "select_player"
	MsgPlayCard = "play_card"
	MsgChooseTarget = "choose_target"

	//server to client messages
	MsgGameState = "game_state"
	MsgHAnd = "hand"
	MsgTargets = "targets"
	MsgError = "error"
)

type Envelope struct {
	Type string `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type SelectPlayerPayload struct {
	PlayerID int `json:"playerId"`
}

type PlayCardPayload struct {
	CardID int `json:"cardId"`
}

type ChooseTargetPayload struct {
	CardID   int `json:"cardId"`
	TargetID int `json:"targetId"`
}

type HandPayload struct {
	PlayerID int   `json:"playerId"`
	CardIDs  []int `json:"cardIds"`
}

type TargetsPayload struct {
	CardID    int   `json:"cardId"`
	TargetIDs []int `json:"targetIds"`
}