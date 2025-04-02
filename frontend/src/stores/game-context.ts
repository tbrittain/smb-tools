import {defineStore} from 'pinia'

type GameContextState = {
  gameId: string | null
  gameTitle: string
}

export const useGameContextStore = defineStore('game-context', {
  state: () => ({
    gameId: null,
    gameTitle: '',
  } as GameContextState),
  actions: {
    setGameContext(gameId: string, gameTitle: string) {
      this.gameId = gameId
      this.gameTitle = gameTitle
    },
    clearGameContext() {
      this.gameId = null
      this.gameTitle = ''
    },
  },
})
