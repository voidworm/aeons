// PlayableCardDTO.effectType of the card kinds
export const EFFECT_PLAYER_MOVE = 0;
export const EFFECT_PLAYER_DISTRACT = 1;
export const EFFECT_PLAYER_ATTACK = 2;
export const EFFECT_PLAYER_HARVEST = 3;

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
}

export interface HarvestNodeDTO {
  id: number;
  name: string;
  flavor: string;
  locationId: number;
  nutrition: HarvestYieldDTO;
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
  fainted: boolean; // the wound made the player faint
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

export interface HarvestResultDTO {
  harvestableId: number;
  nutrition: number;
  resources: number;
}

export interface UpkeepPhaseDTO {
  events: UpkeepEventDTO[];
}
