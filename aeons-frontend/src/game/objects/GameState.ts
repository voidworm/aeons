export interface YieldDTO {
  capacity: number;
  maxCapacity: number;
  yieldAmount: number;
  respawnTicks: number;
  currentTicks: number;
}

export interface LocationDTO {
  id: number;
  name: string;
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
}

export interface HarvestableDTO {
  id: number;
  name: string;
  locationId: number;
  healing: YieldDTO;
  resources: YieldDTO;
}

export interface CardDTO {
  id: number;
  name: string;
  effectText: string;
  effectType: number;
  canBeCast: boolean;
  range: number;
  targetType: number | null;
  candidateIds: number[];
}

export interface GameStateDTO {
  turnCounter: number;
  locations: LocationDTO[];
  players: PlayerDTO[];
  creatures: CreatureDTO[];
  harvestables: HarvestableDTO[];
  cards: CardDTO[];
}
