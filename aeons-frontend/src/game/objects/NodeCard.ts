import type Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';
import { Card } from './Card';
import type { CollectMode, NodeDTO, PoolDTO } from './GameState';

export const poolText = (pool: PoolDTO, name: string) =>
  pool.max > 0 ? `${pool.remaining} ${name}` : `No ${name}`;

export class NodeCard extends Card {
  public modes: CollectMode[];
  public dto: NodeDTO; // latest server state of this node
  private stock!: Phaser.GameObjects.Text;

  constructor(scene: MainScene, x: number, y: number, dto: NodeDTO) {
    super(scene, x, y, {
      id: dto.id,
      name: dto.name,
      flavor: dto.flavor,
      faceColor: 0x777777,
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
  }

  sync(dto: NodeDTO) {
    this.dto = dto;
    this.modes = dto.modes;
    this.stock.setText(NodeCard.stockText(dto));
  }

  static stockText(dto: NodeDTO) {
    return `${poolText(dto.food, 'Food')}    ·    ${poolText(dto.material, 'Material')}`;
  }

  static fromDto(scene: MainScene, x: number, y: number, dto: NodeDTO) {
    return new NodeCard(scene, x, y, dto);
  }
}
