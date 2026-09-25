import { useLayoutEffect, useRef } from "react";
import Phaser from "phaser";
import { startGame } from "../game/main";

export function PhaserGame() {
    const gameRef = useRef<Phaser.Game | null>(null)

    useLayoutEffect(() => {
        if (gameRef.current === null) {
            gameRef.current = startGame('game-container')
        }

        return() => {
            gameRef.current?.destroy(true)
            gameRef.current = null
        }
    }, [])

    return <div id="game-container"></div>
}