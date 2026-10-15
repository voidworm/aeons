import type Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';
import { Card } from './Card';
import type { HarvestNodeDTO, HarvestYieldDTO } from './GameState';

const percent = (n: number, d: number) => (d > 0 ? Math.round((n / d) * 100) : 0);

const yieldText = (y: HarvestYieldDTO, name: string) =>
  y.maxCapacity > 0 ? `${percent(y.capacity, y.maxCapacity)}% ${name}` : `No ${name}`;

export class HarvestableCard extends Card {
  private yields!: Phaser.GameObjects.Text;

  constructor(
    scene: MainScene,
    x: number,
    y: number,
    id: number,
    name: string,
    flavor: string,
    nutrition: HarvestYieldDTO,
    resources: HarvestYieldDTO,
  ) {
    super(scene, x, y, {
      id: id,
      name: name,
      flavor: flavor,
      faceColor: 0x777777,
      artHeight: 136,
      shadow: 'flat',
      italicText: true,
      footerHeight: 19.2,
    });

    this.initFrameDetails(nutrition, resources);
  }

  initFrameDetails(nutrition: HarvestYieldDTO, resources: HarvestYieldDTO) {
    this.yields = this.footer(this.yieldsText(nutrition, resources), 'center');
    this.addFrameDetails(this.yields);
  }

  setYields(nutrition: HarvestYieldDTO, resources: HarvestYieldDTO) {
    this.yields.setText(this.yieldsText(nutrition, resources));
  }

  private yieldsText(nutrition: HarvestYieldDTO, resources: HarvestYieldDTO) {
    return `${yieldText(nutrition, 'Nutrition')}    ·    ${yieldText(resources, 'Material')}`;
  }

  static fromDto(scene: MainScene, x: number, y: number, dto: HarvestNodeDTO) {
    return new HarvestableCard(
      scene,
      x,
      y,
      dto.id,
      dto.name,
      dto.flavor,
      dto.nutrition,
      dto.resources,
    );
  }
}
