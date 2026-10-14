package server

import "encoding/json"

const (
	// client to server messages
	MsgNewGame = "new_game"
	MsgSelectPlayer = "select_player"
	MsgPlayCard = "play_card"
	MsgChooseTarget = "choose_target"
	MsgTriggerCreaturePhase = "trigger_creature_phase"
	MsgTriggerUpkeepPhase = "trigger_upkeep_phase"

	//server to client messages
	MsgInitialGameState = "initial_game_state"
	MsgUpdatedGameState = "updated_game_state"
	MsgHAnd = "hand"
	MsgTargets = "targets"
	MsgCreaturePhase = "creature_phase"
	MsgUpkeepPhase = "upkeep_phase"
	MsgHarvestResult = "harvest_result"
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

type CreaturePhasePayload struct {
	Events []CreatureEventDTO `json:"events"`
}

type CreatureEventDTO struct {
	CreatureID     int                 `json:"creatureId"`
	FromLocationID int                 `json:"fromLocationId"`
	ToLocationID   int                 `json:"toLocationId"`
	Attacks        []CreatureAttackDTO `json:"attacks"`
	Exhausted      bool                `json:"exhausted"`
}

type UpkeepPhasePayload struct {
	Events []UpkeepEventDTO `json:"events"`
}

type UpkeepEventDTO struct {
	CreatureID int `json:"creatureId"`
}

type LogPayload struct {
	Entries []string `json:"entries"`
}

type HarvestResultPayload struct {
	HarvestableID int `json:"harvestableId"`
	Healing       int `json:"healing"`
	Resources     int `json:"resources"`
}

type CreatureAttackDTO struct {
	TargetPlayerID int `json:"targetPlayerId"`
	Damage         int `json:"damage"`
}
