import Phaser from 'phaser'
import { MainScene } from './scenes/MainScene'

const config: Phaser.Types.Core.GameConfig = {
    type: Phaser.AUTO,
    backgroundColor: '#1e1e2e',
    scale : {
        mode: Phaser.Scale.RESIZE,
        width: '100%',
        height: '100%',
    },
    scene: [MainScene]
}

export function startGame(parent: string): Phaser.Game {
    return new Phaser.Game({ ...config, parent})
}