import Phaser from 'phaser';
import { CURSOR_GRAB, CURSOR_INTERACTIVE } from '../cursors';

//moves several containers together by pointer deltas in screen space while a handle is dragged
//the group stays fully on screen
export class DragGroup {
  private readonly scene: Phaser.Scene;
  private readonly members: Phaser.GameObjects.Container[];

  constructor(scene: Phaser.Scene, members: Phaser.GameObjects.Container[]) {
    this.scene = scene;
    this.members = members;
  }

  //the handle must not be interactive yet; being interactive also keeps hovers and clicks from reaching the board underneath
  enableOn(handle: Phaser.GameObjects.Shape): void {
    const last = new Phaser.Math.Vector2();
    handle.setInteractive({ draggable: true, cursor: CURSOR_INTERACTIVE });
    handle.on('dragstart', (pointer: Phaser.Input.Pointer) => {
      last.set(pointer.x, pointer.y);
      this.scene.game.canvas.style.cursor = CURSOR_GRAB;
    });
    handle.on('drag', (pointer: Phaser.Input.Pointer) => {
      const { dx, dy } = this.clampedDelta(pointer.x - last.x, pointer.y - last.y);
      for (const member of this.members) {
        const parentScale = member.parentContainer?.scaleX ?? 1; // positions are local to a scaled parent
        member.setPosition(member.x + dx / parentScale, member.y + dy / parentScale);
      }
      last.set(pointer.x, pointer.y);
    });
    handle.on('dragend', () => {
      this.scene.game.canvas.style.cursor = CURSOR_INTERACTIVE;
    });
  }

  private clampedDelta(dx: number, dy: number): { dx: number; dy: number } {
    const { width, height } = this.scene.scale;
    const bounds = this.members
      .map((member) => member.getBounds())
      .reduce((union, rect) => Phaser.Geom.Rectangle.Union(union, rect, union));
    return {
      dx: Phaser.Math.Clamp(dx, -bounds.left, width - bounds.right),
      dy: Phaser.Math.Clamp(dy, -bounds.top, height - bounds.bottom),
    };
  }
}
