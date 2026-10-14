import type {
  CreaturePhaseDTO,
  GameStateDTO,
  HarvestResultDTO,
  UpkeepPhaseDTO,
} from '../game/objects/GameState';

export type ServerMessage =
  | { type: 'initial_game_state'; payload: GameStateDTO }
  | { type: 'updated_game_state'; payload: GameStateDTO }
  | { type: 'creature_phase'; payload: CreaturePhaseDTO }
  | { type: 'upkeep_phase'; payload: UpkeepPhaseDTO }
  | { type: 'harvest_result'; payload: HarvestResultDTO }
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
