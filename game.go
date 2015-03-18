package unimud

// A Game is an instance of a uniMUD game.
type Game struct {
}

// NewGame creates a new unimud game instance.
func NewGame() *Game {
	return new(Game)
}

// ListenConsole listens for player input on standard input
// and outputs text on standard output.
func (g *Game) ListenConsole() {
	for {
		p := newPlayer(g)
		p.run()
	}
}
