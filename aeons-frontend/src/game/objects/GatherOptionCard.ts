import type { MainScene } from '../scenes/MainScene';
import { Card } from './Card';
import type { CollectMode, NodeDTO } from './GameState';
import { poolText } from './NodeCard';

const LABEL: Record<CollectMode, string> = { harvest: 'Harvest', forage: 'Forage' };

//one half of a node that offers both harvesting and foraging, exists only while a gather card is cast
export class GatherOptionCard extends Card {
  public readonly mode: CollectMode;

  constructor(scene: MainScene, x: number, y: number, dto: NodeDTO, mode: CollectMode) {
    super(scene, x, y, {
      id: dto.id,
      name: dto.name,
      flavor: LABEL[mode],
      faceColor: 0x777777,
      artHeight: 136,
      shadow: 'flat',
      centerText: true,
      footerHeight: 19.2,
    });
    this.mode = mode;
    const pool =
      mode === 'harvest' ? poolText(dto.food, 'Food') : poolText(dto.material, 'Material');
    this.addFrameDetails(this.footer(pool, 'center'));
  }

  protected canBeDragged(): boolean {
    return false;
  }
}
