import Phaser from 'phaser';
import type { MainScene } from '../scenes/MainScene';
import { CARD_FONT_FAMILY } from '../fonts';
import { ControlButton } from './ControlButton';
import type { PaymentDTO } from './GameState';

interface PayerRow {
  playerId: number;
  resources: number;
  amount: number;
  count: Phaser.GameObjects.Text;
  less: ControlButton;
  more: ControlButton;
}

//modal that lets the players split a quest's cost between them; confirming sends the split, cancelling leaves the quest unfinished
export class PaymentPopup extends Phaser.GameObjects.Container {
  private static readonly WIDTH = 560;
  private static readonly ROW_HEIGHT = 64;
  private static readonly HEADER_HEIGHT = 130;
  private static readonly FOOTER_HEIGHT = 100;
  private static readonly COLOR = 0xd8c9a3; // same as the card text boxes
  private static readonly OUTLINE_COLOR = 0x2a1809;
  private static readonly DEPTH = 2000;
  private static readonly STEP_BUTTON_WIDTH = 56;

  private readonly payment: PaymentDTO;
  private rows: PayerRow[] = [];
  private total: Phaser.GameObjects.Text;
  private confirm: ControlButton;

  constructor(
    scene: MainScene,
    payment: PaymentDTO,
    questName: string,
    onConfirm: (payments: Record<number, number>) => void,
    onCancel: () => void,
  ) {
    const { width, height } = scene.scale;
    super(scene, width / 2, height / 2);
    this.payment = payment;
    const { WIDTH, ROW_HEIGHT, HEADER_HEIGHT, FOOTER_HEIGHT, COLOR, OUTLINE_COLOR } = PaymentPopup;
    const panelHeight = HEADER_HEIGHT + payment.payers.length * ROW_HEIGHT + FOOTER_HEIGHT;
    const top = -panelHeight / 2;

    // swallows clicks, so nothing behind the popup reacts
    const dim = new Phaser.GameObjects.Rectangle(scene, 0, 0, width, height, 0x000000, 0.5).setInteractive();
    const panel = new Phaser.GameObjects.Rectangle(scene, 0, 0, WIDTH, panelHeight, COLOR).setStrokeStyle(
      4,
      OUTLINE_COLOR,
    );
    const title = this.text(scene, questName, 0, top + 40, '34px', true);
    this.total = this.text(scene, '', 0, top + 90, '28px');
    this.add([dim, panel, title, this.total]);

    payment.payers.forEach((payer, i) => {
      const y = top + HEADER_HEIGHT + i * ROW_HEIGHT + ROW_HEIGHT / 2;
      const row: PayerRow = {
        playerId: payer.playerId,
        resources: payer.resources,
        amount: 0,
        count: this.text(scene, '0', 60, y, '30px', true),
        less: new ControlButton(scene, '-', () => this.change(row, -1), PaymentPopup.STEP_BUTTON_WIDTH),
        more: new ControlButton(scene, '+', () => this.change(row, 1), PaymentPopup.STEP_BUTTON_WIDTH),
      };
      row.less.setPosition(0, y);
      row.more.setPosition(120, y);
      const name = this.text(scene, `${payer.name} (has ${payer.resources})`, -WIDTH / 2 + 30, y, '28px');
      name.setOrigin(0, 0.5);
      this.add([name, row.count, row.less, row.more]);
      this.rows.push(row);
    });

    const buttonsY = panelHeight / 2 - FOOTER_HEIGHT / 2;
    this.confirm = new ControlButton(scene, 'Confirm', () => onConfirm(this.payments()), 200);
    this.confirm.setPosition(-120, buttonsY);
    const cancel = new ControlButton(scene, 'Cancel', onCancel, 200);
    cancel.setPosition(120, buttonsY);
    this.add([this.confirm, cancel]);

    this.refresh();
    this.setDepth(PaymentPopup.DEPTH);
    scene.add.existing(this);
  }

  private change(row: PayerRow, by: number): void {
    row.amount += by;
    this.refresh();
  }

  private paid(): number {
    return this.rows.reduce((sum, row) => sum + row.amount, 0);
  }

  private payments(): Record<number, number> {
    return Object.fromEntries(this.rows.map((row) => [row.playerId, row.amount]));
  }

  private refresh(): void {
    const paid = this.paid();
    const complete = paid === this.payment.amount;
    this.total.setText(`${paid} / ${this.payment.amount}`);
    for (const row of this.rows) {
      row.count.setText(String(row.amount));
      row.less.setEnabled(row.amount > 0);
      row.more.setEnabled(!complete && row.amount < row.resources);
    }
    this.confirm.setEnabled(complete);
  }

  private text(scene: MainScene, content: string, x: number, y: number, fontSize: string, bold = false) {
    return new Phaser.GameObjects.Text(scene, x, y, content, {
      fontFamily: CARD_FONT_FAMILY,
      fontSize,
      fontStyle: bold ? 'bold' : 'normal',
      color: '#000000',
      align: 'center',
    }).setOrigin(0.5);
  }
}
