// PlayableCardDTO.effectType of the card kinds
export const EFFECT_PLAYER_MOVE = 0;
export const EFFECT_PLAYER_DISTRACT = 1;
export const EFFECT_PLAYER_ATTACK = 2;
export const EFFECT_PLAYER_GATHER = 3;
export const EFFECT_PLAYER_BOND = 4;
export const EFFECT_KINDRED_MOVE = 5;
export const EFFECT_FINISH_QUEST = 6;
export const EFFECT_PLAYER_HIDE = 7;
export const EFFECT_PLAYER_REST = 8;

export type LocationVisibility = 'visible' | 'memory' | 'veiled';

export interface LocationDTO {
  id: number;
  name: string;
  flavor: string;
  connections: number[];
  visibility: LocationVisibility; // memory: unveiled once, veiled now; veiled: never unveiled
}

export interface PlayerDTO {
  id: number;
  name: string;
  locationId: number;
  wounds: number;
  maxWounds: number;
  fainted: boolean;
  hidden: boolean;
  actions: number;
  hand: number[];
  food: number;
  resources: number;
  energy: number;
  maxEnergy: number;
}

export interface CreatureDTO {
  id: number;
  name: string;
  identNames: string[];
  locationId: number;
  distress: number;
  maxDistress: number;
  size: number;
  flavor: string;
  exhausted: boolean;
  hidden: boolean;
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
  candidateIds: number[]; // the targets the server accepts, filled for cards that target quests or creatures
}

export interface GameStateDTO {
  cycle: number;
  locations: LocationDTO[];
  players: PlayerDTO[];
  creatures: CreatureDTO[];
  nodes: NodeDTO[];
  cardsInHand: PlayableCardDTO[];
  scenarioName: string;
  quests: QuestDTO[];
  pendingPayment: PaymentDTO | null;
  activationOrder: ActivationDTO[]; // the creatures in the order they activate next, only filled while the veil is lifted
}

export interface ActivationDTO {
  creatureId: number;
  name: string;
}

export interface GoalDTO {
  text: string;
  met: boolean;
}

export interface QuestDTO {
  id: number;
  name: string;
  flavor: string;
  goals: GoalDTO[];
  mandatory: boolean;
  finishable: boolean;
}

export interface PayerDTO {
  playerId: number;
  name: string;
  resources: number;
}

//a finished quest waits until the players have said who pays how much
export interface PaymentDTO {
  questId: number;
  amount: number;
  payers: PayerDTO[];
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
  ready: boolean; // the creature readied during this event, before it moves
}

export interface WorldPhaseDTO {
  events: CreatureEventDTO[];
}

export interface GatherResultDTO {
  nodeId: number;
  mode: CollectMode;
}
