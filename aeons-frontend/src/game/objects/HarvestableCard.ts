import type { MainScene } from '../scenes/MainScene';
import { Card } from './Card';
import type { HarvestableDTO, YieldDTO } from './GameState';

const percent = (n: number, d: number) => (d > 0 ? Math.round((n / d) * 100) : 0);

export class HarvestableCard extends Card {
  constructor(
    scene: MainScene,
    x: number,
    y: number,
    name: string,
    flavor: string,
    healing: YieldDTO,
    resources: YieldDTO,
  ) {
    super(scene, x, y, {
      name: name,
      flavor: flavor,
      faceColor: 0x777777,
      artHeight: 170,
      shadow: 'flat',
      italicText: true,
      footerHeight: 24,
    });

    this.initFrameDetails(healing, resources);
  }

  initFrameDetails(healing: YieldDTO, resources: YieldDTO) {
    const food = percent(healing.capacity, healing.maxCapacity);
    const material = percent(resources.capacity, resources.maxCapacity);
    this.addFrameDetails(this.footer(`${food}% Food    ·    ${material}% Material`, 'center'));
  }

  static fromDto(scene: MainScene, x: number, y: number, dto: HarvestableDTO) {
    return new HarvestableCard(scene, x, y, dto.name, dto.flavor, dto.healing, dto.resources);
  }
}
