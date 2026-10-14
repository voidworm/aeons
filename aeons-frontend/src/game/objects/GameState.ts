// PlayableCardDTO.effectType of the card kinds
export const EFFECT_PLAYER_MOVE = 0;
export const EFFECT_PLAYER_DISTRACT = 1;
export const EFFECT_PLAYER_ATTACK = 2;

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
  health: number;
  maxHealth: number;
  actions: number;
  hand: number[];
  resources: number;
  energy: number;
  damage: number;
}

export interface CreatureDTO {
  id: number;
  name: string;
  locationId: number;
  health: number;
  maxHealth: number;
  damage: number;
  flavor: string;
  exhausted: boolean;
}

export interface HarvestNodeDTO {
  id: number;
  name: string;
  flavor: string;
  locationId: number;
  healing: HarvestYieldDTO;
  resources: HarvestYieldDTO;
}

export interface HarvestYieldDTO {
  capacity: number;
  maxCapacity: number;
  yieldAmount: number;
  respawnTicks: number;
  currentTicks: number;
}

export interface PlayableCardDTO {
  id: number;
  name: string;
  effectText: string;
  effectType: number;
  cost: integer;
  canBeCast: boolean;
  range: number;
  hand: number[];
  targetType: number | null;
}

export interface GameStateDTO {
  turnCounter: number;
  locations: LocationDTO[];
  players: PlayerDTO[];
  creatures: CreatureDTO[];
  harvestNodes: HarvestNodeDTO[];
  cardsInHand: PlayableCardDTO[];
}

export interface CreatureAttackDTO {
  targetPlayerId: number;
  damage: number;
}

export interface CreatureEventDTO {
  creatureId: number;
  fromLocationId: number;
  toLocationId: number;
  attacks: CreatureAttackDTO[];
  exhausted: boolean; // the creature became exhausted during this event
}

export interface CreaturePhaseDTO {
  events: CreatureEventDTO[];
}

export interface UpkeepEventDTO {
  creatureId: number; // the creature recovers from exhaustion
}

export interface UpkeepPhaseDTO {
  events: UpkeepEventDTO[];
}
