import type { GameStateDTO } from '../game/objects/GameState';

export type ServerMessage =
  | { type: 'game_state'; payload: GameStateDTO }
  | { type: 'poc_reply'; payload: string }
  | { type: 'poc_done'; payload: string }
  | { type: 'error'; payload: string };

export function connect(onMessage: (msg: ServerMessage) => void): WebSocket {
  const ws = new WebSocket('ws://localhost:8080/ws');

  ws.onmessage = (event: MessageEvent<string>) => {
    onMessage(JSON.parse(event.data) as ServerMessage);
  };

  return ws;
}
