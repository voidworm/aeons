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
			gs.Init()
			s.send(ctx, conn, MsgGameState, Snapshot(gs))
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
