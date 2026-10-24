import Phaser from 'phaser';

interface Word {
  text: Phaser.GameObjects.Text;
  width: number;
  startsLine: boolean; // follows a line break in the markup
}

//centred, word-wrapped text where *asterisks* mark bold passages and a newline starts a new line
export class RichText extends Phaser.GameObjects.Container {
  constructor(
    scene: Phaser.Scene,
    x: number,
    y: number,
    markup: string,
    style: Phaser.Types.GameObjects.Text.TextStyle,
    wrapWidth: number,
    anchorY: number, // 0 top, 0.5 middle, 1 bottom of the text block at y
  ) {
    super(scene, x, y);
    const words = this.words(markup, style);
    const spaceWidth = this.measure(scene, ' ', style);
    const lines = this.wrap(words, wrapWidth, spaceWidth);
    const lineHeight = words[0]?.text.height ?? 0;
    const top = -lines.length * lineHeight * anchorY;

    lines.forEach((line, row) => {
      const lineWidth = line.reduce((sum, w) => sum + w.width, 0) + spaceWidth * (line.length - 1);
      let cursor = -lineWidth / 2;
      for (const word of line) {
        word.text.setPosition(cursor, top + row * lineHeight);
        cursor += word.width + spaceWidth;
      }
    });
  }

  private words(markup: string, style: Phaser.Types.GameObjects.Text.TextStyle): Word[] {
    let breakPending = false;
    return markup.split('*').flatMap((part, i) =>
      part.split('\n').flatMap((line, row) => {
        if (row > 0) breakPending = true;
        return line
          .split(/\s+/)
          .filter(Boolean)
          .map((content) => {
            const text = new Phaser.GameObjects.Text(this.scene, 0, 0, content, {
              ...style,
              fontStyle: i % 2 === 1 ? 'bold' : style.fontStyle,
            });
            this.add(text);
            const word = { text, width: text.width, startsLine: breakPending };
            breakPending = false;
            return word;
          });
      }),
    );
  }

  private wrap(words: Word[], wrapWidth: number, spaceWidth: number): Word[][] {
    const lines: Word[][] = [[]];
    let lineWidth = 0;
    for (const word of words) {
      const current = lines[lines.length - 1];
      const widthWithWord = lineWidth + (current.length > 0 ? spaceWidth : 0) + word.width;
      if (current.length > 0 && (word.startsLine || widthWithWord > wrapWidth)) {
        lines.push([word]);
        lineWidth = word.width;
      } else {
        current.push(word);
        lineWidth = widthWithWord;
      }
    }
    return lines;
  }

  private measure(
    scene: Phaser.Scene,
    content: string,
    style: Phaser.Types.GameObjects.Text.TextStyle,
  ) {
    const probe = new Phaser.GameObjects.Text(scene, 0, 0, content, style);
    const { width } = probe;
    probe.destroy();
    return width;
  }
}
