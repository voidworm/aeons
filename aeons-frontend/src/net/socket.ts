import type {
  AwakeningPhaseDTO,
  GameStateDTO,
  GatherResultDTO,
  Phase,
  WorldPhaseDTO,
} from '../game/objects/GameState';

export type ServerMessage =
  | { type: 'initial_game_state'; payload: GameStateDTO }
  | { type: 'updated_game_state'; payload: GameStateDTO }
  | { type: 'phase_changed'; payload: { phase: Phase; cycle: number } }
  | { type: 'world_phase'; payload: WorldPhaseDTO }
  | { type: 'awakening_phase'; payload: AwakeningPhaseDTO }
  | { type: 'gather_result'; payload: GatherResultDTO }
  | { type: 'log'; payload: { entries: string[] } }
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
