// PlayableCardDTO.effectType of the card kinds
export const EFFECT_PLAYER_MOVE = 0;
export const EFFECT_PLAYER_DISTRACT = 1;
export const EFFECT_PLAYER_ATTACK = 2;
export const EFFECT_PLAYER_GATHER = 3;
export const EFFECT_PLAYER_BOND = 4;
export const EFFECT_KINDRED_MOVE = 5;

export type Phase = 'spirit' | 'world' | 'awakening';

export interface LocationDTO {
  id: number;
  name: string;
  flavor: string;
  connections: number[];
}

export interface PlayerDTO {
  id: number;
  name: string;
  locationId: number;
  wounds: number;
  maxWounds: number;
  fainted: boolean;
  actions: number;
  hand: number[];
  food: number;
  resources: number;
  energy: number;
}

export interface CreatureDTO {
  id: number;
  name: string;
  locationId: number;
  health: number;
  maxHealth: number;
  flavor: string;
  exhausted: boolean;
  bonded: boolean;
  kindredOf: number | null; // id of the player the creature is kindred to
}

export interface NodeDTO {
  id: number;
  name: string;
  flavor: string;
  locationId: number;
  food: PoolDTO;
  material: PoolDTO;
  modes: CollectMode[]; // what can be collected here right now
}

export type CollectMode = 'harvest' | 'forage';

export interface PoolDTO {
  remaining: number;
  max: number;
  regrowthTimer: number; // 0 means it never grows back
  turnsToRegrow: number;
}

export interface PlayableCardDTO {
  id: number;
  name: string;
  effectText: string;
  flavor: string;
  effectType: number;
  cost: integer;
  canBeCast: boolean;
  range: number;
  hand: number[];
  targetType: number | null;
}

export interface GameStateDTO {
  cycle: number;
  phase: Phase;
  locations: LocationDTO[];
  players: PlayerDTO[];
  creatures: CreatureDTO[];
  nodes: NodeDTO[];
  cardsInHand: PlayableCardDTO[];
}

export interface CreatureAttackDTO {
  targetPlayerId: number;
  fainted: boolean; // the wound made the player faint
}

export interface CreatureEventDTO {
  creatureId: number;
  fromLocationId: number;
  toLocationId: number;
  attacks: CreatureAttackDTO[];
  exhausted: boolean; // the creature became exhausted during this event
}

export interface WorldPhaseDTO {
  events: CreatureEventDTO[];
}

export interface AwakeningEventDTO {
  creatureId: number; // the creature recovers from exhaustion
}

export interface GatherResultDTO {
  nodeId: number;
  mode: CollectMode;
}

export interface AwakeningPhaseDTO {
  events: AwakeningEventDTO[];
}
