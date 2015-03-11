package unimud

// A Game is an instance of a uniMUD game.
type Game struct {
}

func NewGame() *Game {
	return new(Game)
}

func (g *Game) ListenConsole() {
	for {
		p := newPlayer(g)
		p.run()
	}
}
