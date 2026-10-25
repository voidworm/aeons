package server

import "encoding/json"

const (
	// client to server messages
	MsgNewGame = "new_game"
	MsgSelectPlayer = "select_player"
	MsgPlayCard = "play_card"
	MsgChooseTarget = "choose_target"
	MsgEndSpiritPhase = "end_spirit_phase"
	MsgStartAwakeningPhase = "start_awakening_phase"
	MsgStartSpiritPhase = "start_spirit_phase"
	MsgTriggerWorldPhase = "trigger_world_phase" // playtest, runs the phase without changing the game's phase
	MsgTriggerAwakeningPhase = "trigger_awakening_phase" // playtest, runs the phase without changing the game's phase

	//server to client messages
	MsgInitialGameState = "initial_game_state"
	MsgUpdatedGameState = "updated_game_state"
	MsgHAnd = "hand"
	MsgTargets = "targets"
	MsgPhaseChanged = "phase_changed"
	MsgWorldPhase = "world_phase"
	MsgAwakeningPhase = "awakening_phase"
	MsgGatherResult = "gather_result"
	MsgLog = "log"
	MsgError = "error"

	//poc messages
	//fe->be
	MsgPocHello = "poc_hello"
	MsgPocAck = "poc_ack"

	//be->fe
	MsgPocReply = "poc_reply"
	MsgPocDone = "poc_done"
)

type Envelope struct {
	Type string `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type SelectPlayerPayload struct {
	PlayerID int `json:"playerId"`
}

type PlayCardPayload struct {
	CardID   int `json:"cardId"`
	TargetID int `json:"targetId"`
	Mode     string `json:"mode"` // harvest or forage, only read for gather cards
}

type ChooseTargetPayload struct {
	CardID   int `json:"cardId"`
	TargetID int `json:"targetId"`
	Mode     string `json:"mode"` // harvest or forage, only read for gather cards
}

type HandPayload struct {
	PlayerID int   `json:"playerId"`
	CardIDs  []int `json:"cardIds"`
}

type TargetsPayload struct {
	CardID    int   `json:"cardId"`
	TargetIDs []int `json:"targetIds"`
}

type PhaseChangedPayload struct {
	Phase string `json:"phase"`
	Cycle int    `json:"cycle"`
}

type WorldPhasePayload struct {
	Events []CreatureEventDTO `json:"events"`
}

type CreatureEventDTO struct {
	CreatureID     int                 `json:"creatureId"`
	FromLocationID int                 `json:"fromLocationId"`
	ToLocationID   int                 `json:"toLocationId"`
	Attacks        []CreatureAttackDTO `json:"attacks"`
	Exhausted      bool                `json:"exhausted"`
}

type AwakeningPhasePayload struct {
	Events []AwakeningEventDTO `json:"events"`
}

type AwakeningEventDTO struct {
	CreatureID int `json:"creatureId"`
}

type LogPayload struct {
	Entries []string `json:"entries"`
}

type GatherResultPayload struct {
	NodeID int    `json:"nodeId"`
	Mode   string `json:"mode"`
}

type CreatureAttackDTO struct {
	TargetPlayerID int  `json:"targetPlayerId"`
	Fainted        bool `json:"fainted"`
}
