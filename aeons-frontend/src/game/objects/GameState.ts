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
