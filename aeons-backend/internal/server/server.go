package server

import (
	"aeons/internal/game"
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type Server struct{}

func New() *Server {
	return &Server{}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Connection request from %s", r.RemoteAddr)

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"localhost:5173"},
	})
	if err != nil {
		log.Printf("caught error: %v", err)
		return
	}
	defer conn.CloseNow()

	log.Printf("client connected: %s", r.RemoteAddr)
	s.readLoop(r.Context(), conn)

	log.Printf("client disconnected: %s", r.RemoteAddr)
}

func (s *Server) readLoop(ctx context.Context, conn *websocket.Conn) {
	var gs *game.GameState

	for {
		var msg Envelope
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			return
		}

		switch msg.Type {
		case MsgNewGame:
			gs = &game.GameState{}
			log.Print("Received new game request, initalising....")
			gs.Init()
			s.send(ctx, conn, MsgInitialGameState, Snapshot(gs))
		case MsgTriggerCreaturePhase:
			if gs == nil {
				s.send(ctx, conn, MsgError, "no game running")
				continue
			}
			s.send(ctx, conn, MsgCreaturePhase, creaturePhasePayload(gs.ResolveCreaturePhase()))
			s.send(ctx, conn, MsgUpdatedGameState, Snapshot(gs))
		case MsgTriggerUpkeepPhase:
			if gs == nil {
				s.send(ctx, conn, MsgError, "no game running")
				continue
			}
			s.send(ctx, conn, MsgUpkeepPhase, upkeepPhasePayload(gs.ResolveUpkeepPhase()))
			s.send(ctx, conn, MsgUpdatedGameState, Snapshot(gs))
		case MsgPocHello:
			log.Print("Received a hello!")
			s.send(ctx,conn, MsgPocReply, "...")
		case MsgPocAck:
			log.Print("Received an acknowledgement!")
			s.send(ctx,conn, MsgPocDone, "...")
		default:
			s.send(ctx, conn, MsgError, "unknown message type: "+msg.Type)
			
		}
	}
}

func (s *Server) send(ctx context.Context, conn *websocket.Conn, typ string, payload any) {
	raw, err := json.Marshal(payload)
	if err != nil {
		log.Printf("marshal %s: %v", typ, err)
		return
	}
	if err := wsjson.Write(ctx, conn, Envelope{Type: typ, Payload: raw}); err != nil {
		log.Printf("write %s: %v", typ, err)
	}
}

func creaturePhasePayload(events []game.CreatureEvent) CreaturePhasePayload {
	out := CreaturePhasePayload{Events: make([]CreatureEventDTO, 0, len(events))}
	for _, e := range events {
		attacks := make([]CreatureAttackDTO, 0, len(e.Attacks))
		for _, a := range e.Attacks {
			attacks = append(attacks, CreatureAttackDTO{TargetPlayerID: a.TargetPlayerID, Damage: a.Damage})
		}
		out.Events = append(out.Events, CreatureEventDTO{
			CreatureID: e.CreatureID, FromLocationID: e.FromLocationID,
			ToLocationID: e.ToLocationID, Attacks: attacks, Exhausted: e.Exhausted,
		})
	}
	return out
}

func upkeepPhasePayload(events []game.UpkeepEvent) UpkeepPhasePayload {
	out := UpkeepPhasePayload{Events: make([]UpkeepEventDTO, 0, len(events))}
	for _, e := range events {
		out.Events = append(out.Events, UpkeepEventDTO{CreatureID: e.CreatureID})
	}
	return out
}
