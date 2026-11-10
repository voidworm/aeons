import type Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';
import { Card } from './Card';
import type { CollectMode, NodeDTO, PoolDTO } from './GameState';

export const poolText = (pool: PoolDTO, name: string) =>
  pool.max > 0 ? `${pool.remaining} ${name}` : `No ${name}`;

export const FOOD_COLOR = 0xc4551b; // deep orange
export const MATERIAL_COLOR = 0x3d9a42; // lush green
const NO_STOCK_COLOR = 0x777777;

// a node is coloured by what it can hold, not by what is left
function faceOf(dto: NodeDTO): { faceColor: number; faceColorEnd?: number } {
  const hasFood = dto.food.max > 0;
  const hasMaterial = dto.material.max > 0;
  if (hasFood && hasMaterial) return { faceColor: FOOD_COLOR, faceColorEnd: MATERIAL_COLOR };
  if (hasFood) return { faceColor: FOOD_COLOR };
  if (hasMaterial) return { faceColor: MATERIAL_COLOR };
  return { faceColor: NO_STOCK_COLOR };
}

export class NodeCard extends Card {
  public modes: CollectMode[];
  public dto: NodeDTO; // latest server state of this node
  private stock!: Phaser.GameObjects.Text;
  private veil: Phaser.GameObjects.Graphics; // darkens a gathering spot at a veiled location

  constructor(scene: MainScene, x: number, y: number, dto: NodeDTO) {
    super(scene, x, y, {
      id: dto.id,
      name: dto.name,
      flavor: dto.flavor,
      ...faceOf(dto),
      artHeight: 136,
      shadow: 'flat',
      italicText: true,
      centerTitle: true,
      footerHeight: 19.2,
    });
    this.dto = dto;
    this.modes = dto.modes;
    this.stock = this.footer(NodeCard.stockText(dto), 'center');
    this.addFrameDetails(this.stock);
    this.veil = this.roundedRect(0, 0, this.cardWidth, this.cardHeight, Card.CORNER_RADIUS, 0x111111, 0.45).setVisible(false);
    this.surface.add(this.veil);
  }

  setVeiled(veiled: boolean) {
    this.veil.setVisible(veiled);
  }

  sync(dto: NodeDTO) {
    this.dto = dto;
    this.modes = dto.modes;
    this.stock.setText(NodeCard.stockText(dto));
  }

  static stockText(dto: NodeDTO) {
    const hasFood = dto.food.max > 0;
    const hasMaterial = dto.material.max > 0;
    if (hasFood && !hasMaterial) return poolText(dto.food, 'Food');
    if (hasMaterial && !hasFood) return poolText(dto.material, 'Material');
    return `${poolText(dto.food, 'Food')}    ·    ${poolText(dto.material, 'Material')}`;
  }

  static fromDto(scene: MainScene, x: number, y: number, dto: NodeDTO) {
    return new NodeCard(scene, x, y, dto);
  }
}
