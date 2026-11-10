package server

import "encoding/json"

const (
	// client to server messages
	MsgNewGame = "new_game"
	MsgSelectPlayer = "select_player"
	MsgPlayCard = "play_card"
	MsgChooseTarget = "choose_target"
	MsgPayForQuest = "pay_for_quest"
	MsgCancelPayment = "cancel_payment"
	MsgAdvanceWorld = "advance_world" // the next creature takes its turn, without a player action
	MsgSetFogLifted = "set_fog_lifted" // playtest, shows everything regardless of what the players know

	//server to client messages
	MsgInitialGameState = "initial_game_state"
	MsgUpdatedGameState = "updated_game_state"
	MsgHAnd = "hand"
	MsgTargets = "targets"
	MsgWorldPhase = "world_phase"
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

type PayForQuestPayload struct {
	QuestID  int         `json:"questId"`
	Payments map[int]int `json:"payments"` // player id to the resources that player pays
}

type SetFogLiftedPayload struct {
	Lifted bool `json:"lifted"`
}

type HandPayload struct {
	PlayerID int   `json:"playerId"`
	CardIDs  []int `json:"cardIds"`
}

type TargetsPayload struct {
	CardID    int   `json:"cardId"`
	TargetIDs []int `json:"targetIds"`
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
	Ready          bool                `json:"ready"`
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
