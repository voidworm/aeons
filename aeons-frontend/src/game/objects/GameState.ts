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

export interface HarvestableDTO {
  id: number;
  name: string;
  flavor: string;
  locationId: number;
  healing: YieldDTO;
  resources: YieldDTO;
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
  harvestables: HarvestableDTO[];
  playables: PlayableCardDTO[];
}
